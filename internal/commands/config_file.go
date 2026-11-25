package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/internal/yamledit"
)

// configName is the name of the config in the project root (ADR 0002).
const configName = "goro.yaml"

// configFile is a goro.yaml that a run writes besides the code: the copy init saves in the
// project root, or the goro.yaml that add changed.
type configFile struct {
	// path is where the file goes.
	path string
	// shown names the file in the output: its path relative to the work dir when it is inside.
	shown string
	// action is ActionCreate, ActionOverwrite or ActionUnchanged.
	action  Action
	current []byte
	content []byte
}

// planned returns f as an entry of a plan, for printing.
func (f *configFile) planned() plannedFile {
	return plannedFile{path: f.shown, action: f.action, rendered: f.content, current: f.current}
}

func (f *configFile) write() error {
	if f.action != ActionCreate && f.action != ActionOverwrite {
		return nil
	}

	return os.WriteFile(f.path, f.content, 0o644)
}

// shownPath returns path relative to workDir, slash-separated, when it is inside workDir, and
// path as it is otherwise.
func shownPath(path, workDir string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	absDir, err := filepath.Abs(workDir)
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(absDir, absPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return path
	}

	return filepath.ToSlash(rel)
}

// initConfigFile returns the goro.yaml that init leaves in the project root (ADR 0002), or nil
// when there is nothing to write:
//
//   - for a config from outside the project, a copy in the project root with work_dir: . and
//     the app name cfg holds when the config has none (see rootGoroYaml). A different goro.yaml
//     in the project root is an error: init never overwrites a file;
//   - for the project's own goro.yaml, that file with what LoadConfig had to ask for filled in:
//     work_dir: . and the app name; nil when it gives both;
//   - nil for a config that wasn't read from a file.
func initConfigFile(cfg config.Config) (*configFile, error) {
	if cfg.Path == "" {
		return nil, nil
	}

	target := filepath.Join(cfg.App.WorkDir, configName)
	same, err := sameFile(cfg.Path, target)
	if err != nil {
		return nil, err
	}

	source, err := os.ReadFile(cfg.Path)
	if err != nil {
		return nil, err
	}
	content, err := rootGoroYaml(source, cfg.Path, cfg.App.Name, !same)
	if err != nil {
		if same {
			return nil, fmt.Errorf("init: can't fill in work_dir and name in %s: %w", cfg.Path, err)
		}
		return nil, fmt.Errorf("init: can't save a copy of %s in the project root: %w", cfg.Path, err)
	}

	if same {
		if bytes.Equal(content, source) {
			return nil, nil
		}
		return &configFile{path: target, shown: configName, action: ActionOverwrite, current: source, content: content}, nil
	}

	f := &configFile{path: target, shown: configName, action: ActionCreate, content: content}
	current, err := os.ReadFile(target)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return f, nil
	case err != nil:
		return nil, err
	case !bytes.Equal(current, content):
		return nil, fmt.Errorf("init: the project root holds a different goro.yaml, %s, than the copy of %s init would save there; "+
			"move it away, or run goro init --config %s to start from it", target, cfg.Path, target)
	}
	f.action = ActionUnchanged
	f.current = current

	return f, nil
}

// sameFile reports whether a and b name the same file; b may not exist.
func sameFile(a, b string) (bool, error) {
	ai, err := os.Stat(a)
	if err != nil {
		return false, err
	}
	bi, err := os.Stat(b)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return os.SameFile(ai, bi), nil
}

// rootGoroYaml returns source, the goro.yaml at path, as init leaves it in the project root:
// app.name set to name when source has none, which is the answer to LoadConfig's question, and
// app.work_dir set to "." when source has none, or always for a copy from outside the project.
// Each is set in place, or inserted after the last key of app; every other byte stays.
func rootGoroYaml(source []byte, path, name string, outside bool) ([]byte, error) {
	want, err := config.ParseConfig(source, path)
	if err != nil {
		return nil, err
	}
	setName := want.App.Name == ""
	setWorkDir := outside || want.App.WorkDir == ""
	if !setName && !setWorkDir {
		return source, nil
	}

	doc, err := yamledit.Parse(source)
	if err != nil {
		return nil, err
	}
	_, app := yamledit.Lookup(doc.Root(), "app")
	if app == nil || app.Kind != yaml.MappingNode {
		return nil, errors.New("it has no app mapping")
	}
	if yamledit.IsFlow(app) {
		return nil, errors.New("app is a flow mapping, such as {name: demo}; write it as a block mapping, one key per line")
	}

	if setName {
		if err := doc.SetScalar(app, "name", name); err != nil {
			return nil, fmt.Errorf("set app.name: %w", err)
		}
		want.App.Name = name
	}
	if setWorkDir {
		if err := doc.SetScalar(app, "work_dir", "."); err != nil {
			return nil, fmt.Errorf("set app.work_dir: %w", err)
		}
		want.App.WorkDir = filepath.Dir(path)
	}
	content := doc.Bytes()

	return content, checkEdit(content, path, want)
}

// checkEdit fails unless content, the goro.yaml at path after text edits, says what want says:
// the edits must change nothing but what they set out to.
func checkEdit(content []byte, path string, want config.Config) error {
	got, err := config.ParseConfig(content, path)
	if err == nil && !reflect.DeepEqual(configParts(got), configParts(want)) {
		err = errors.New("the edited file says something else")
	}
	if err != nil {
		return fmt.Errorf("goro can't make this change by inserting lines (%v); make it by hand", err)
	}

	return nil
}

// configParts returns the parts of cfg that goro.yaml holds, with empty lists as nil.
func configParts(cfg config.Config) config.Config {
	list := func(s []string) []string {
		if len(s) == 0 {
			return nil
		}
		return append([]string(nil), s...)
	}

	res := config.Config{App: cfg.App}
	for _, s := range cfg.Storages {
		res.Storages = append(res.Storages, s)
	}
	res.UseCase = config.UseCase{Methods: list(cfg.UseCase.Methods), Deps: list(cfg.UseCase.Deps)}
	for _, s := range cfg.Services {
		res.Services = append(res.Services, config.Service{Name: s.Name, Methods: list(s.Methods), Deps: list(s.Deps)})
	}
	for _, a := range cfg.Adapters {
		res.Adapters = append(res.Adapters, config.Adapter{Name: a.Name, Storage: a.Storage, Methods: list(a.Methods)})
	}

	return res
}
