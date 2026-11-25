// Package yamledit changes a YAML file by inserting lines into its text, so every line the file
// had stays as it was, with its comments, blank lines, quoting and order. yaml.v3 only locates the
// parts of the file (their lines, columns and styles); the inserted lines follow the file's own
// indentation. Changing a value in place is the one exception, for SetScalar.
package yamledit

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// ErrFlow is returned for an edit inside a flow collection, such as [a, b] or {a: b}: it can't
// grow by inserting a line.
var ErrFlow = errors.New("a flow collection can't grow by a line")

// Value is what an edit inserts: a string, which becomes a scalar, a List or a Map.
type Value interface{}

// List is a block list of values: strings or Maps.
type List []Value

// Map is a block mapping, with its keys in order. A Field whose value is an empty List or Map is
// left out.
type Map []Field

// Field is one key of a Map and its value.
type Field struct {
	Key   string
	Value Value
}

// Document is the text of a YAML file and the yaml.v3 nodes that locate its parts. Edits are
// queued against the original text, so the nodes stay valid, and Bytes applies them.
type Document struct {
	lines []string
	// eol ends every inserted line: the file's own line ending.
	eol    string
	root   *yaml.Node
	layout layout

	inserts      []insertion
	replacements []replacement
}

// layout is how the file indents what it nests.
type layout struct {
	// mapIndent is how far the keys of a nested block mapping sit right of their parent key.
	mapIndent int
	// seqIndent is how far the dashes of a block list sit right of its key; 0 for lists at the
	// column of their key.
	seqIndent int
	// itemIndent is how far the first key of a mapping in a block list sits right of its dash.
	itemIndent int
}

type insertion struct {
	// after is the 1-based line the text goes after.
	after int
	// col is the indentation of the first line that isn't blank.
	col   int
	lines []string
}

type replacement struct {
	// line is 1-based; start and end are byte offsets in it.
	line, start, end int
	text             string
}

// Parse reads data, one YAML document.
func Parse(data []byte) (*Document, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	d := &Document{lines: splitLines(string(data)), eol: "\n"}
	if len(d.lines) > 0 && strings.HasSuffix(d.lines[0], "\r\n") {
		d.eol = "\r\n"
	}
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		d.root = doc.Content[0]
	}
	d.layout = d.findLayout()

	return d, nil
}

// Root returns the top node of the document, usually a mapping, or nil for an empty document.
func (d *Document) Root() *yaml.Node {
	return d.root
}

// Lookup returns the key and value nodes of key in the mapping m, or nils when m has no such key
// or isn't a mapping. For a key m holds twice it returns the last one, which is the one that
// counts for a decoder that allows duplicates, such as yaml.v2.
func Lookup(m *yaml.Node, key string) (k, v *yaml.Node) {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil, nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			k, v = m.Content[i], m.Content[i+1]
		}
	}

	return k, v
}

// IsEmpty reports whether n is the value of a key written without one, as in "methods:".
func IsEmpty(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode && n.Tag == "!!null" && n.Value == ""
}

// IsFlow reports whether n is a flow collection, such as [a, b] or {a: b}.
func IsFlow(n *yaml.Node) bool {
	return n.Style&yaml.FlowStyle != 0
}

// AppendItems inserts items after the last item of the block list seq, at its indentation.
func (d *Document) AppendItems(seq *yaml.Node, items ...Value) error {
	if seq.Kind != yaml.SequenceNode {
		return fmt.Errorf("line %d: not a list", seq.Line)
	}
	if IsFlow(seq) || len(seq.Content) == 0 {
		return ErrFlow
	}

	dash := d.dashColumn(seq)
	lines, err := d.renderItems(items, dash)
	if err != nil {
		return err
	}
	d.insert(d.end(seq.Content[len(seq.Content)-1], dash), lines)

	return nil
}

// InsertField inserts key with v at the end of the block mapping m, at the column of its keys.
// A new key of the root mapping gets a blank line before it when the file puts blank lines
// between its root keys.
func (d *Document) InsertField(m *yaml.Node, key string, v Value) error {
	if m.Kind != yaml.MappingNode {
		return fmt.Errorf("line %d: not a mapping", m.Line)
	}
	if IsFlow(m) || len(m.Content) == 0 {
		return ErrFlow
	}

	col := keyColumn(m)
	lines, err := d.renderField(key, v, col)
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		return fmt.Errorf("nothing to insert for %s", key)
	}
	if m == d.root && d.rootKeysApart() {
		lines = append([]string{d.eol}, lines...)
	}
	d.insert(d.end(m, col), lines)

	return nil
}

// Fill gives the key node key, written without a value (see IsEmpty), the value v: a List or a
// Map, on the lines after the key.
func (d *Document) Fill(key *yaml.Node, v Value) error {
	col := key.Column - 1
	var lines []string
	var err error
	switch v := v.(type) {
	case List:
		lines, err = d.renderItems(v, col+d.layout.seqIndent)
	case Map:
		lines, err = d.renderFields(v, col+d.layout.mapIndent)
	default:
		return fmt.Errorf("line %d: %s can only take a list or a mapping", key.Line, key.Value)
	}
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		return fmt.Errorf("nothing to insert for %s", key.Value)
	}
	d.insert(d.extend(key.Line, col), lines)

	return nil
}

// SetScalar sets key of the block mapping m to value. It replaces the scalar value on the key's
// line, fills in a key written without a value, or inserts the key at the end of m. It refuses a
// value with an anchor or a tag.
func (d *Document) SetScalar(m *yaml.Node, key, value string) error {
	if m.Kind != yaml.MappingNode {
		return fmt.Errorf("line %d: not a mapping", m.Line)
	}
	if IsFlow(m) {
		return ErrFlow
	}

	k, v := Lookup(m, key)
	if k == nil {
		return d.InsertField(m, key, value)
	}
	if v.Kind != yaml.ScalarNode {
		return fmt.Errorf("line %d: %s isn't a scalar", v.Line, key)
	}
	if decorated(v) {
		return fmt.Errorf("line %d: the value of %s has an anchor or a tag, which an alias may share, so it can't change in place; "+
			"remove them or set the value by hand", v.Line, key)
	}

	text := scalarText(value)
	if IsEmpty(v) {
		_, keyEnd, err := d.span(k)
		if err != nil {
			return err
		}
		line := d.lines[k.Line-1]
		colon := keyEnd + len(line[keyEnd:]) - len(strings.TrimLeft(line[keyEnd:], " "))
		if colon >= len(line) || line[colon] != ':' {
			return fmt.Errorf("line %d: no colon after %s", k.Line, key)
		}
		d.replacements = append(d.replacements, replacement{line: k.Line, start: colon + 1, end: colon + 1, text: " " + text})
		return nil
	}

	start, end, err := d.span(v)
	if err != nil {
		return err
	}
	if d.lines[v.Line-1][start:end] != text {
		d.replacements = append(d.replacements, replacement{line: v.Line, start: start, end: end, text: text})
	}

	return nil
}

// Bytes returns the text with every queued edit applied.
func (d *Document) Bytes() []byte {
	lines := append([]string(nil), d.lines...)

	reps := append([]replacement(nil), d.replacements...)
	sort.SliceStable(reps, func(i, j int) bool {
		if reps[i].line != reps[j].line {
			return reps[i].line < reps[j].line
		}
		return reps[i].start > reps[j].start
	})
	for _, r := range reps {
		l := lines[r.line-1]
		lines[r.line-1] = l[:r.start] + r.text + l[r.end:]
	}

	// After the same line, lines that sit further right belong inside the node that ends there,
	// so they go before lines for the nodes around it.
	inserts := append([]insertion(nil), d.inserts...)
	sort.SliceStable(inserts, func(i, j int) bool {
		if inserts[i].after != inserts[j].after {
			return inserts[i].after < inserts[j].after
		}
		return inserts[i].col > inserts[j].col
	})

	var b strings.Builder
	next := 0
	emit := func(after int) {
		for ; next < len(inserts) && inserts[next].after == after; next++ {
			for _, l := range inserts[next].lines {
				b.WriteString(l)
			}
		}
	}
	emit(0)
	for i, l := range lines {
		b.WriteString(l)
		if next < len(inserts) && inserts[next].after == i+1 && !strings.HasSuffix(l, "\n") {
			b.WriteString(d.eol)
		}
		emit(i + 1)
	}

	return []byte(b.String())
}

func (d *Document) insert(after int, lines []string) {
	col := 0
	for _, l := range lines {
		if !isBlank(l) {
			col = indentOf(l)
			break
		}
	}
	d.inserts = append(d.inserts, insertion{after: after, col: col, lines: lines})
}

// end returns the last line of the text of n, which starts at column col: the line of its last
// token, or a line below it that still belongs to n (see extend).
func (d *Document) end(n *yaml.Node, col int) int {
	return d.extend(lastLine(n), col)
}

// extend returns line, or the last line after it that is indented more than col with only blank
// lines and other such lines between: the rest of a multi-line value, or comments inside the node
// that ends on line. A trailing comment at col or left of it, such as a commented-out list item,
// stays after the inserted lines.
func (d *Document) extend(line, col int) int {
	end := line
	for i := line + 1; i <= len(d.lines); i++ {
		l := d.lines[i-1]
		if isBlank(l) {
			continue
		}
		if indentOf(l) <= col {
			break
		}
		end = i
	}

	return end
}

// dashColumn returns the column of the dashes of the block list seq. yaml.v3 puts a list with an
// anchor or a tag at that anchor or tag, so for one the column comes from the first line after it
// that starts with a dash.
func (d *Document) dashColumn(seq *yaml.Node) int {
	if !decorated(seq) {
		return seq.Column - 1
	}

	last := seq.Line
	if len(seq.Content) > 0 {
		last = seq.Content[0].Line
	}
	for i := seq.Line + 1; i <= last && i <= len(d.lines); i++ {
		if l := d.lines[i-1]; strings.HasPrefix(strings.TrimLeft(l, " "), "-") {
			return indentOf(l)
		}
	}

	return seq.Column - 1
}

// keyColumn returns the column of the keys of the block mapping m: that of its first key, since
// yaml.v3 puts a mapping with an anchor or a tag at that anchor or tag.
func keyColumn(m *yaml.Node) int {
	if len(m.Content) > 0 {
		return m.Content[0].Column - 1
	}

	return m.Column - 1
}

// decorated reports whether n has an anchor or an explicit tag in the text.
func decorated(n *yaml.Node) bool {
	return n.Anchor != "" || n.Style&yaml.TaggedStyle != 0
}

// lastLine returns the line of the last token of n.
func lastLine(n *yaml.Node) int {
	res := n.Line
	for _, c := range n.Content {
		if l := lastLine(c); l > res {
			res = l
		}
	}

	return res
}

// rootKeysApart reports whether the file puts a blank line before a key of its root mapping, other
// than the first, above the comments of that key.
func (d *Document) rootKeysApart() bool {
	if d.root == nil || d.root.Kind != yaml.MappingNode {
		return false
	}
	for i := 2; i < len(d.root.Content); i += 2 {
		j := d.root.Content[i].Line - 1
		for j >= 1 && isComment(d.lines[j-1]) {
			j--
		}
		if j >= 1 && isBlank(d.lines[j-1]) {
			return true
		}
	}

	return false
}

// findLayout reads the indentation of the first nested block mapping, block list under a key and
// mapping in a block list. What the file doesn't show gets 2.
func (d *Document) findLayout() layout {
	l := layout{mapIndent: -1, seqIndent: -1, itemIndent: -1}

	var walk func(n *yaml.Node)
	walk = func(n *yaml.Node) {
		if IsFlow(n) {
			return
		}
		switch n.Kind {
		case yaml.MappingNode:
			for i := 0; i+1 < len(n.Content); i += 2 {
				k, v := n.Content[i], n.Content[i+1]
				if !IsFlow(v) && len(v.Content) > 0 && v.Content[0].Line > k.Line {
					switch {
					case v.Kind == yaml.MappingNode && l.mapIndent < 0:
						l.mapIndent = keyColumn(v) - (k.Column - 1)
					case v.Kind == yaml.SequenceNode && l.seqIndent < 0:
						l.seqIndent = d.dashColumn(v) - (k.Column - 1)
					}
				}
				walk(v)
			}
		case yaml.SequenceNode:
			dash := d.dashColumn(n)
			for _, item := range n.Content {
				if item.Kind == yaml.MappingNode && !IsFlow(item) && l.itemIndent < 0 && d.dashBefore(item, dash) {
					l.itemIndent = keyColumn(item) - dash
				}
				walk(item)
			}
		}
	}
	if d.root != nil {
		walk(d.root)
	}

	if l.mapIndent < 1 {
		l.mapIndent = 2
	}
	if l.seqIndent < 0 {
		l.seqIndent = 2
	}
	if l.itemIndent < 2 {
		l.itemIndent = 2
	}

	return l
}

// dashBefore reports whether the line of item has the dash of its list at column dash, so the
// item starts on the line of its dash.
func (d *Document) dashBefore(item *yaml.Node, dash int) bool {
	line := d.lines[item.Line-1]

	return indentOf(line) == dash && strings.HasPrefix(line[dash:], "-")
}

func (d *Document) renderField(key string, v Value, col int) ([]string, error) {
	head := strings.Repeat(" ", col) + scalarText(key) + ":"

	switch v := v.(type) {
	case string:
		return []string{head + " " + scalarText(v) + d.eol}, nil
	case List:
		if len(v) == 0 {
			return nil, nil
		}
		items, err := d.renderItems(v, col+d.layout.seqIndent)
		if err != nil {
			return nil, err
		}
		return append([]string{head + d.eol}, items...), nil
	case Map:
		fields, err := d.renderFields(v, col+d.layout.mapIndent)
		if err != nil || len(fields) == 0 {
			return nil, err
		}
		return append([]string{head + d.eol}, fields...), nil
	default:
		return nil, fmt.Errorf("can't insert a %T for %s", v, key)
	}
}

func (d *Document) renderFields(m Map, col int) ([]string, error) {
	res := make([]string, 0, len(m))
	for _, f := range m {
		lines, err := d.renderField(f.Key, f.Value, col)
		if err != nil {
			return nil, err
		}
		res = append(res, lines...)
	}

	return res, nil
}

func (d *Document) renderItems(items []Value, dash int) ([]string, error) {
	res := make([]string, 0, len(items))
	for _, item := range items {
		switch item := item.(type) {
		case string:
			res = append(res, strings.Repeat(" ", dash)+"- "+scalarText(item)+d.eol)
		case Map:
			col := dash + d.layout.itemIndent
			lines, err := d.renderFields(item, col)
			if err != nil {
				return nil, err
			}
			if len(lines) == 0 {
				return nil, errors.New("can't insert an empty mapping as a list item")
			}
			lines[0] = strings.Repeat(" ", dash) + "-" + strings.Repeat(" ", d.layout.itemIndent-1) + lines[0][col:]
			res = append(res, lines...)
		default:
			return nil, fmt.Errorf("can't insert a %T as a list item", item)
		}
	}

	return res, nil
}

// span returns the byte offsets of the scalar n on its line: from its first character to after
// its last, closing quotes included. It fails for a scalar that goes on to the next line.
func (d *Document) span(n *yaml.Node) (start, end int, err error) {
	line := strings.TrimRight(d.lines[n.Line-1], "\r\n")
	start = byteOffset(line, n.Column-1)
	rest := line[start:]

	length := -1
	switch {
	case n.Style&yaml.DoubleQuotedStyle != 0:
		length = closingQuote(rest, '"')
	case n.Style&yaml.SingleQuotedStyle != 0:
		length = closingQuote(rest, '\'')
	case n.Style&(yaml.LiteralStyle|yaml.FoldedStyle) == 0 && strings.HasPrefix(rest, n.Value):
		length = len(n.Value)
	}
	if length < 0 {
		return 0, 0, fmt.Errorf("line %d: %q doesn't end on its line", n.Line, n.Value)
	}

	return start, start + length, nil
}

// closingQuote returns the length of the quoted scalar that s starts with, up to and including
// its closing quote, or -1 when it doesn't close in s.
func closingQuote(s string, quote byte) int {
	for i := 1; i < len(s); i++ {
		switch {
		case quote == '"' && s[i] == '\\':
			i++
		case s[i] == quote && quote == '\'' && i+1 < len(s) && s[i+1] == '\'':
			i++
		case s[i] == quote:
			return i + 1
		}
	}

	return -1
}

// scalarText writes s as a YAML scalar on one line: plain when that reads back as the same
// string, quoted otherwise.
func scalarText(s string) string {
	out, err := yaml.Marshal(s)
	text := strings.TrimSuffix(string(out), "\n")
	if err == nil && !strings.Contains(text, "\n") {
		return text
	}

	out, _ = yaml.Marshal(&yaml.Node{Kind: yaml.ScalarNode, Style: yaml.DoubleQuotedStyle, Tag: "!!str", Value: s})

	return strings.TrimSuffix(string(out), "\n")
}

// byteOffset returns the offset of the rune at index col in line, or len(line) past its end.
func byteOffset(line string, col int) int {
	offset := 0
	for i := 0; i < col && offset < len(line); i++ {
		_, size := utf8.DecodeRuneInString(line[offset:])
		offset += size
	}

	return offset
}

// splitLines splits text after each newline; the last line may lack one.
func splitLines(text string) []string {
	if text == "" {
		return nil
	}
	lines := strings.SplitAfter(text, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}

func indentOf(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func isBlank(line string) bool {
	return strings.TrimSpace(line) == ""
}

func isComment(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "#")
}
