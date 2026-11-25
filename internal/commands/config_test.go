package commands_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/hanagantig/goro/internal/commands"
	"github.com/hanagantig/goro/internal/config"
)

// Rules under test: ADR 0002, goro.yaml in the project root is the source of truth. The default
// --config path is covered in the cmd package, which runs the command line.

// sameDir reports whether a and b name the same existing directory. A relative path counts from
// the test's working directory, which is never a project dir here.
func sameDir(a, b string) bool {
	ai, err := os.Stat(a)
	if err != nil {
		return false
	}
	bi, err := os.Stat(b)
	if err != nil {
		return false
	}

	return os.SameFile(ai, bi)
}

func mkdir(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o775); err != nil {
		t.Fatal(err)
	}
}

func TestLoadConfigResolvesRelativeWorkDir(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	for _, d := range []string{project, filepath.Join(project, "app"), filepath.Join(root, "other"), filepath.Join(root, "abs")} {
		mkdir(t, d)
	}
	path := filepath.Join(project, "goro.yaml")

	cases := []struct {
		name, workDir, want string
	}{
		{name: "dot is the directory of goro.yaml", workDir: ".", want: project},
		{name: "a relative path counts from the directory of goro.yaml", workDir: "app", want: filepath.Join(project, "app")},
		{name: "a path up from the directory of goro.yaml", workDir: "../other", want: filepath.Join(root, "other")},
		{name: "an absolute path is kept", workDir: filepath.Join(root, "abs"), want: filepath.Join(root, "abs")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			yaml := fmt.Sprintf("app:\n  name: demo\n  module: demo\n  work_dir: %s\n", c.workDir)
			if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
				t.Fatal(err)
			}

			cfg, err := commands.LoadConfig(path)
			if err != nil {
				t.Fatalf("load config: %v", err)
			}
			if !sameDir(cfg.App.WorkDir, c.want) {
				t.Errorf("work_dir %q in %s: LoadConfig gives work dir %q, want %s", c.workDir, path, cfg.App.WorkDir, c.want)
			}
		})
	}
}

func TestInitSavesGoroYamlInProjectRoot(t *testing.T) {
	t.Run("a config from outside the project is copied with work_dir: .", func(t *testing.T) {
		// testdata/init/goro.yaml lives in root/templates and has work_dir: ../app.
		root := t.TempDir()
		project := filepath.Join(root, "app")
		mkdir(t, project)
		source := filepath.Join(root, "templates", "goro.yaml")
		writeFile(t, root, "templates/goro.yaml", readTestdata(t, "init/goro.yaml"))

		cfg, err := commands.LoadConfig(source)
		if err != nil {
			t.Fatalf("load config: %v", err)
		}
		// Init writes into the work dir, so stop before it unless work_dir: ../app resolved
		// to the project.
		if !sameDir(cfg.App.WorkDir, project) {
			t.Fatalf("LoadConfig gives work dir %q for work_dir: ../app in %s, want %s", cfg.App.WorkDir, source, project)
		}

		if err := commands.Init(cfg, commands.Options{SkipGoMod: true, Out: io.Discard}); err != nil {
			t.Fatalf("init: %v", err)
		}

		saved := filepath.Join(project, "goro.yaml")
		if !exists(project, "goro.yaml") {
			t.Fatalf("init didn't save goro.yaml in the project root")
		}
		// The copy is the source with work_dir: ., comments and key order kept, which is
		// testdata/add/goro.yaml.
		assertGolden(t, saved, "add/goro.yaml")
		if got := readFile(t, root, "templates/goro.yaml"); got != readTestdata(t, "init/goro.yaml") {
			t.Errorf("init changed the source config:\n%s", got)
		}

		// The saved goro.yaml describes the project: update from it has nothing to do.
		cfg, err = commands.LoadConfig(saved)
		if err != nil {
			t.Fatalf("load the saved goro.yaml: %v", err)
		}
		if !sameDir(cfg.App.WorkDir, project) {
			t.Fatalf("LoadConfig gives work dir %q for the saved goro.yaml, want %s", cfg.App.WorkDir, project)
		}
		report, err := commands.Update(cfg, commands.Options{SkipGoMod: true, Out: io.Discard})
		if err != nil {
			t.Fatalf("update from the saved goro.yaml: %v", err)
		}
		for _, f := range report.Files {
			if f.Action == commands.ActionCreate || f.Action == commands.ActionOverwrite {
				t.Errorf("update from the saved goro.yaml reports %q for %s", f.Action, f.Path)
			}
		}
	})

	t.Run("a work dir the config doesn't give is saved as work_dir: .", func(t *testing.T) {
		// newConfig loads baseConfig, which has no work_dir, from outside the project and sets
		// the work dir itself, as the answer to init's question does.
		dir := t.TempDir()
		cfg := newConfig(t, dir)
		initProject(t, cfg)

		if !exists(dir, "goro.yaml") {
			t.Fatalf("init didn't save goro.yaml in the project root")
		}
		saved, err := config.NewConfig(filepath.Join(dir, "goro.yaml"))
		if err != nil {
			t.Fatalf("read the saved goro.yaml: %v", err)
		}
		if saved.App.WorkDir != "." {
			t.Errorf("saved goro.yaml has work_dir %q, want .", saved.App.WorkDir)
		}
		// %v prints nil and empty lists alike, which a saved file may choose between.
		for _, part := range []struct {
			name      string
			got, want interface{}
		}{
			{"app.name", saved.App.Name, cfg.App.Name},
			{"app.module", saved.App.Module, cfg.App.Module},
			{"storages", saved.Storages, cfg.Storages},
			{"use_case", saved.UseCase, cfg.UseCase},
			{"services", saved.Services, cfg.Services},
			{"adapters", saved.Adapters, cfg.Adapters},
		} {
			if got, want := fmt.Sprintf("%+v", part.got), fmt.Sprintf("%+v", part.want); got != want {
				t.Errorf("saved goro.yaml %s = %s, want %s", part.name, got, want)
			}
		}
	})
}
