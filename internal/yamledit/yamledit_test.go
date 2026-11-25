package yamledit

import (
	"errors"
	"testing"

	"gopkg.in/yaml.v3"
)

func parse(t *testing.T, text string) *Document {
	t.Helper()

	d, err := Parse([]byte(text))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	return d
}

// node walks keys from the root mapping and returns the value at the end.
func node(t *testing.T, d *Document, keys ...string) *yaml.Node {
	t.Helper()

	n := d.Root()
	for _, k := range keys {
		_, n = Lookup(n, k)
		if n == nil {
			t.Fatalf("no %s", k)
		}
	}

	return n
}

func keyNode(t *testing.T, d *Document, parent *yaml.Node, key string) *yaml.Node {
	t.Helper()

	k, _ := Lookup(parent, key)
	if k == nil {
		t.Fatalf("no key %s", key)
	}

	return k
}

func assertText(t *testing.T, d *Document, want string) {
	t.Helper()

	got := string(d.Bytes())
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
	var n yaml.Node
	if err := yaml.Unmarshal([]byte(got), &n); err != nil {
		t.Errorf("the result isn't YAML: %v", err)
	}
}

func check(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}

func TestAppendItems(t *testing.T) {
	cases := []struct {
		name, in, want string
		keys           []string
		items          []Value
	}{
		{
			name:  "before a trailing commented-out item, at the list's indentation",
			in:    "storages:\n  - mysql\n  # - pgsqlx\n\nnext: 1\n",
			keys:  []string{"storages"},
			items: []Value{"pgsqlx"},
			want:  "storages:\n  - mysql\n  - pgsqlx\n  # - pgsqlx\n\nnext: 1\n",
		},
		{
			name:  "a list at the column of its key",
			in:    "a:\n  b:\n  - x\n  c: 1\n",
			keys:  []string{"a", "b"},
			items: []Value{"z"},
			want:  "a:\n  b:\n  - x\n  - z\n  c: 1\n",
		},
		{
			name:  "after a comment inside the last item",
			in:    "l:\n  - name: A\n    m:\n      - X\n      # - Y\n  # - name: B\n",
			keys:  []string{"l"},
			items: []Value{Map{{"name", "C"}}},
			want:  "l:\n  - name: A\n    m:\n      - X\n      # - Y\n  - name: C\n  # - name: B\n",
		},
		{
			name:  "after a multi-line value",
			in:    "l:\n  - |\n    one\n    two\n\nnext: 1\n",
			keys:  []string{"l"},
			items: []Value{"three"},
			want:  "l:\n  - |\n    one\n    two\n  - three\n\nnext: 1\n",
		},
		{
			name:  "a last line without a newline",
			in:    "l:\n  - a",
			keys:  []string{"l"},
			items: []Value{"b"},
			want:  "l:\n  - a\n  - b\n",
		},
		{
			name:  "mappings in the file's layout",
			in:    "l:\n-   name: A\n    m:\n    - X\n",
			keys:  []string{"l"},
			items: []Value{Map{{"name", "B"}, {"m", List{"W", "Z"}}, {"empty", List{}}}},
			want:  "l:\n-   name: A\n    m:\n    - X\n-   name: B\n    m:\n    - W\n    - Z\n",
		},
		{
			name:  "scalars that need quotes",
			in:    "l:\n  - a\n",
			keys:  []string{"l"},
			items: []Value{"yes", "Y", "12", "a: b", ""},
			want:  "l:\n  - a\n  - \"yes\"\n  - \"Y\"\n  - \"12\"\n  - 'a: b'\n  - \"\"\n",
		},
		{
			name:  "CRLF line endings",
			in:    "l:\r\n  - a\r\n",
			keys:  []string{"l"},
			items: []Value{"b"},
			want:  "l:\r\n  - a\r\n  - b\r\n",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := parse(t, c.in)
			check(t, d.AppendItems(node(t, d, c.keys...), c.items...))
			assertText(t, d, c.want)
		})
	}
}

func TestInsertField(t *testing.T) {
	t.Run("a root key after blank-line separated sections, before the closing comment", func(t *testing.T) {
		d := parse(t, "# head\na: 1\n\n# about b\nb:\n  - x\n  # - y\n\n# The end.\n")
		check(t, d.InsertField(d.Root(), "c", Map{{"m", List{"Z"}}}))
		assertText(t, d, "# head\na: 1\n\n# about b\nb:\n  - x\n  # - y\n\nc:\n  m:\n    - Z\n\n# The end.\n")
	})

	t.Run("a root key without blank lines between sections", func(t *testing.T) {
		d := parse(t, "a:\n  x: 1\nb: 2\n")
		check(t, d.InsertField(d.Root(), "c", List{"Z"}))
		assertText(t, d, "a:\n  x: 1\nb: 2\nc:\n  - Z\n")
	})

	t.Run("a key in a mapping that is a list item", func(t *testing.T) {
		d := parse(t, "l:\n  - name: A\n  - name: B\n")
		item := node(t, d, "l").Content[0]
		check(t, d.InsertField(item, "m", List{"X"}))
		assertText(t, d, "l:\n  - name: A\n    m:\n      - X\n  - name: B\n")
	})

	t.Run("two inserts after the same line keep their order", func(t *testing.T) {
		d := parse(t, "a: 1\nl:\n  - x\n")
		check(t, d.AppendItems(node(t, d, "l"), "z"))
		check(t, d.InsertField(d.Root(), "b", "two"))
		assertText(t, d, "a: 1\nl:\n  - x\n  - z\nb: two\n")
	})

	t.Run("a flow mapping", func(t *testing.T) {
		d := parse(t, "a: {x: 1}\n")
		if err := d.InsertField(node(t, d, "a"), "z", "two"); !errors.Is(err, ErrFlow) {
			t.Errorf("error = %v, want ErrFlow", err)
		}
	})
}

func TestAppendItemsToFlowList(t *testing.T) {
	for _, in := range []string{"l: [a, b]\n", "l: []\n"} {
		d := parse(t, in)
		if err := d.AppendItems(node(t, d, "l"), "c"); !errors.Is(err, ErrFlow) {
			t.Errorf("%q: error = %v, want ErrFlow", in, err)
		}
	}
}

func TestFill(t *testing.T) {
	t.Run("a list", func(t *testing.T) {
		d := parse(t, "a:\n  m:\n\nb: 1\n")
		check(t, d.Fill(keyNode(t, d, node(t, d, "a"), "m"), List{"X", "W"}))
		assertText(t, d, "a:\n  m:\n    - X\n    - W\n\nb: 1\n")
	})

	t.Run("a mapping, after a comment", func(t *testing.T) {
		d := parse(t, "a: 1\nb: # later\n  # none yet\nc: 2\n")
		check(t, d.Fill(keyNode(t, d, d.Root(), "b"), Map{{"m", List{"X"}}}))
		assertText(t, d, "a: 1\nb: # later\n  # none yet\n  m:\n    - X\nc: 2\n")
	})
}

func TestSetScalar(t *testing.T) {
	cases := []struct {
		name, in, key, value, want string
	}{
		{"a plain value with a comment", "app:\n  dir: ../app # here\n", "dir", ".", "app:\n  dir: . # here\n"},
		{"a double-quoted value", "app:\n  dir: \"a \\\" b\" # here\n", "dir", ".", "app:\n  dir: . # here\n"},
		{"a single-quoted value", "app:\n  dir: 'it''s'\n", "dir", ".", "app:\n  dir: .\n"},
		{"an empty value", "app:\n  dir: \"\"\n", "dir", ".", "app:\n  dir: .\n"},
		{"a key without a value", "app:\n  dir: # here\n", "dir", ".", "app:\n  dir: . # here\n"},
		{"the same value", "app:\n  dir: .  # here\n", "dir", ".", "app:\n  dir: .  # here\n"},
		{"a missing key", "app:\n  name: x\n\nnext: 1\n", "dir", ".", "app:\n  name: x\n  dir: .\n\nnext: 1\n"},
		{"a value that needs quotes", "app:\n  name:\n", "name", "no", "app:\n  name: \"no\"\n"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := parse(t, c.in)
			check(t, d.SetScalar(node(t, d, "app"), c.key, c.value))
			assertText(t, d, c.want)
		})
	}

	t.Run("a value that goes on to the next line", func(t *testing.T) {
		d := parse(t, "app:\n  dir: a\n    b\n")
		if err := d.SetScalar(node(t, d, "app"), "dir", "."); err == nil {
			t.Errorf("SetScalar on a multi-line value succeeded:\n%s", d.Bytes())
		}
	})
}

// TestAnchorsAndTags: the node of an anchored or tagged collection starts at its anchor or tag,
// not at its first item or key, so the column of what goes in comes from the items' own lines.
func TestAnchorsAndTags(t *testing.T) {
	t.Run("items of an anchored list", func(t *testing.T) {
		d := parse(t, "use_case:\n  methods: &ucm\n    - SignIn\n    - Pong\nx: 1\n")
		check(t, d.AppendItems(node(t, d, "use_case", "methods"), "SignOut"))
		assertText(t, d, "use_case:\n  methods: &ucm\n    - SignIn\n    - Pong\n    - SignOut\nx: 1\n")
	})

	t.Run("items of a tagged list with tagged items", func(t *testing.T) {
		d := parse(t, "use_case:\n  methods: !!seq\n    - !!str SignIn\n    - Pong\nx: 1\n")
		check(t, d.AppendItems(node(t, d, "use_case", "methods"), "SignOut"))
		assertText(t, d, "use_case:\n  methods: !!seq\n    - !!str SignIn\n    - Pong\n    - SignOut\nx: 1\n")
	})

	t.Run("items of an anchored list with a comment after the anchor", func(t *testing.T) {
		d := parse(t, "use_case:\n  methods: &ucm # c\n    - SignIn\n")
		check(t, d.AppendItems(node(t, d, "use_case", "methods"), "SignOut"))
		assertText(t, d, "use_case:\n  methods: &ucm # c\n    - SignIn\n    - SignOut\n")
	})

	t.Run("a mapping in an anchored list", func(t *testing.T) {
		d := parse(t, "l: &s\n  -   name: A\n")
		check(t, d.AppendItems(node(t, d, "l"), Map{{"name", "B"}}))
		assertText(t, d, "l: &s\n  -   name: A\n  -   name: B\n")
	})

	t.Run("a key of an anchored mapping", func(t *testing.T) {
		d := parse(t, "app: &a\n  name: demo\n  module: x\n")
		check(t, d.SetScalar(node(t, d, "app"), "work_dir", "."))
		assertText(t, d, "app: &a\n  name: demo\n  module: x\n  work_dir: .\n")
	})

	t.Run("the layout of anchored and tagged values", func(t *testing.T) {
		d := parse(t, "a: &x\n    b: 1\nc: !!seq\n- z\n")
		check(t, d.InsertField(d.Root(), "extra", Map{{"k", List{"v"}}}))
		assertText(t, d, "a: &x\n    b: 1\nc: !!seq\n- z\nextra:\n    k:\n    - v\n")
	})
}
