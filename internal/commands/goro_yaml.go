package commands

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/internal/yamledit"
)

// goroYaml is a project's goro.yaml as goro add changes it: the config it holds and its text. An
// add changes both, the config in memory and the text by inserting lines (ADR 0002), and save
// checks that they agree before it writes anything.
type goroYaml struct {
	path string
	// data is the file as add read it.
	data []byte
	// cfg is the config as the add leaves it.
	cfg  config.Config
	doc  *yamledit.Document
	root *yaml.Node
}

// readGoroYaml reads the goro.yaml at path for an add. add never asks for anything, so a
// goro.yaml without app.name or app.work_dir is an error.
func readGoroYaml(path string) (*goroYaml, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%s doesn't exist; run goro add in the project root, where init saves goro.yaml, or point --config to it", path)
	}
	if err != nil {
		return nil, err
	}

	cfg, err := config.ParseConfig(data, path)
	if err != nil {
		return nil, err
	}
	switch {
	case cfg.App.Name == "":
		return nil, fmt.Errorf("%s has no app.name; set it there, goro add doesn't ask for it", path)
	case cfg.App.WorkDir == "":
		return nil, fmt.Errorf("%s has no app.work_dir; set it there, for example to . for the directory that holds it", path)
	}

	doc, err := yamledit.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	root := doc.Root()
	if root == nil || root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s doesn't hold a mapping", path)
	}
	if yamledit.IsFlow(root) {
		return nil, flowMappingError(path, "the whole file")
	}

	return &goroYaml{path: path, data: data, cfg: cfg, doc: doc, root: root}, nil
}

// save validates g.cfg, runs edit to queue the text edits that make the same change, checks that
// the edited text says what g.cfg says, and then writes goro.yaml and runs the update. added names
// the addition, such as service 'Billing', for errors after goro.yaml is written.
func (g *goroYaml) save(opts Options, added string, edit func() error) (Report, error) {
	if err := g.cfg.Validate(); err != nil {
		return Report{}, err
	}
	if err := edit(); err != nil {
		return Report{}, err
	}

	content := g.doc.Bytes()
	if err := checkEdit(content, g.path, g.cfg); err != nil {
		return Report{}, fmt.Errorf("%s: %w", g.path, err)
	}

	conf := &configFile{
		path:    g.path,
		shown:   shownPath(g.path, g.cfg.App.WorkDir),
		action:  ActionOverwrite,
		current: g.data,
		content: content,
		added:   added,
	}

	return run(g.cfg, opts, "add", false, conf)
}

// appendTo queues the insertion of items into the block list at keys under the block mapping m,
// with the keys that m lacks. label is the path of m in goro.yaml, such as adapters.UserRepo, for
// errors; "" for the root.
func (g *goroYaml) appendTo(m *yaml.Node, label string, keys []string, items ...yamledit.Value) error {
	key := keys[0]
	path := key
	if label != "" {
		path = label + "." + key
	}

	// value is what goes under key when goro.yaml lacks it or gives it no value.
	value := func() yamledit.Value {
		var v yamledit.Value = yamledit.List(items)
		for i := len(keys) - 1; i >= 1; i-- {
			v = yamledit.Map{{Key: keys[i], Value: v}}
		}
		return v
	}

	k, v := yamledit.Lookup(m, key)
	switch {
	case k == nil:
		return g.doc.InsertField(m, key, value())
	case yamledit.IsEmpty(v):
		return g.doc.Fill(k, value())
	case len(keys) > 1:
		if v.Kind != yaml.MappingNode {
			return fmt.Errorf("%s: %s isn't a mapping", g.path, path)
		}
		if yamledit.IsFlow(v) {
			return flowMappingError(g.path, path)
		}
		return g.appendTo(v, path, keys[1:], items...)
	case v.Kind != yaml.SequenceNode:
		return fmt.Errorf("%s: %s isn't a list", g.path, path)
	case yamledit.IsFlow(v):
		return flowListError(g.path, path)
	}

	return g.doc.AppendItems(v, items...)
}

// item returns the mapping of the service or adapter called name in section, services or
// adapters, and its path for errors, such as adapters.UserRepo.
func (g *goroYaml) item(section, name string) (*yaml.Node, string, error) {
	label := section + "." + name

	_, seq := yamledit.Lookup(g.root, section)
	if seq == nil || seq.Kind != yaml.SequenceNode {
		return nil, "", fmt.Errorf("%s: %s isn't a list", g.path, section)
	}
	if yamledit.IsFlow(seq) {
		return nil, "", flowListError(g.path, section)
	}

	var found *yaml.Node
	for _, it := range seq.Content {
		if _, n := yamledit.Lookup(it, "name"); n != nil && n.Value == name {
			found = it
		}
	}
	if found == nil {
		return nil, "", fmt.Errorf("%s: %s lists no '%s'", g.path, section, name)
	}
	if yamledit.IsFlow(found) {
		return nil, "", flowMappingError(g.path, label)
	}

	return found, label, nil
}

func flowListError(file, path string) error {
	return fmt.Errorf("%s: %s is a flow list, such as [a, b]; goro add only inserts lines, so write it as a block list, one \"- item\" per line",
		file, path)
}

func flowMappingError(file, path string) error {
	return fmt.Errorf("%s: %s is a flow mapping, such as {name: a}; goro add only inserts lines, so write it as a block mapping, one key per line",
		file, path)
}
