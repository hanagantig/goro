package commands_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

func TestConfigRecordsItsPath(t *testing.T) {
	t.Chdir(t.TempDir())
	path := filepath.Join(t.TempDir(), "goro.yaml")
	writeFile(t, filepath.Dir(path), "goro.yaml", "app:\n  name: demo\n  module: demo\n  work_dir: .\n")

	cfg, err := config.NewConfig(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if cfg.Path != path {
		t.Errorf("config.NewConfig(%q) gives Path %q, want the path it read", path, cfg.Path)
	}

	cfg, err = commands.LoadConfig(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Path != path {
		t.Errorf("LoadConfig(%q) gives Path %q, want the path it read", path, cfg.Path)
	}
}

func TestLoadConfigResolvesRelativeWorkDir(t *testing.T) {
	t.Chdir(t.TempDir())
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

// workDirLine is the work_dir line of testdata/add/goro.yaml.
const workDirLine = "  work_dir: . # the project, relative to this file\n"

// outsideProject lays out root/templates/goro.yaml, holding source, and an empty project dir
// root/app, after moving the test into an empty working directory. It returns the source path and
// the project dir.
func outsideProject(t *testing.T, source string) (sourcePath, project string) {
	t.Helper()

	t.Chdir(t.TempDir())
	root := t.TempDir()
	project = filepath.Join(root, "app")
	mkdir(t, project)
	writeFile(t, root, "templates/goro.yaml", source)

	return filepath.Join(root, "templates", "goro.yaml"), project
}

// loadOutsideConfig loads the config at source and checks that its work dir is project: init
// writes into the work dir, so the test stops before init otherwise.
func loadOutsideConfig(t *testing.T, source, project string) config.Config {
	t.Helper()

	cfg, err := commands.LoadConfig(source)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !sameDir(cfg.App.WorkDir, project) {
		t.Fatalf("LoadConfig gives work dir %q for %s, want %s", cfg.App.WorkDir, source, project)
	}

	return cfg
}

// readWithWorkDir reads the config at source and sets its work dir to workDir itself. That stands
// for the answer to init's question when the config gives no work dir, which LoadConfig would ask,
// and keeps a test from depending on how LoadConfig resolves work_dir.
func readWithWorkDir(t *testing.T, source, workDir string) config.Config {
	t.Helper()

	cfg, err := config.NewConfig(source)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	cfg.App.WorkDir = workDir

	return cfg
}

func TestInitSavesGoroYamlInProjectRoot(t *testing.T) {
	t.Run("a config from outside the project is copied with work_dir: .", func(t *testing.T) {
		// testdata/init/goro.yaml has work_dir: ../app.
		source, project := outsideProject(t, readTestdata(t, "init/goro.yaml"))
		cfg := loadOutsideConfig(t, source, project)

		if err := commands.Init(cfg, commands.Options{SkipGoMod: true, Out: io.Discard}); err != nil {
			t.Fatalf("init: %v", err)
		}

		saved := filepath.Join(project, "goro.yaml")
		if !exists(project, "goro.yaml") {
			t.Fatalf("init didn't save goro.yaml in the project root")
		}
		// The copy is the source with only the work_dir line changed, which is
		// testdata/add/goro.yaml.
		assertGolden(t, saved, "add/goro.yaml")
		if got := readFile(t, filepath.Dir(source), "goro.yaml"); got != readTestdata(t, "init/goro.yaml") {
			t.Errorf("init changed the source config:\n%s", got)
		}

		// The saved goro.yaml describes the project: update from it has nothing to do.
		cfg = loadOutsideConfig(t, saved, project)
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

	t.Run("a config without work_dir gets a work_dir: . line", func(t *testing.T) {
		base := readTestdata(t, "add/goro.yaml")
		src := strings.Replace(base, workDirLine, "", 1)
		source, project := outsideProject(t, src)

		if err := commands.Init(readWithWorkDir(t, source, project), commands.Options{SkipGoMod: true, Out: io.Discard}); err != nil {
			t.Fatalf("init: %v", err)
		}

		if !exists(project, "goro.yaml") {
			t.Fatalf("init didn't save goro.yaml in the project root")
		}
		// The source with one line inserted after the last key of app, at its indentation.
		want := strings.Replace(base, workDirLine, "  work_dir: .\n", 1)
		if got := readFile(t, project, "goro.yaml"); got != want {
			t.Errorf("saved goro.yaml:\n%s\nwant the source with a work_dir: . line inserted:\n%s", got, want)
		}
		assertOnlyInsertedLines(t, src, filepath.Join(project, "goro.yaml"))
	})

	t.Run(`a config with work_dir: "" gets only that line changed`, func(t *testing.T) {
		// The form example/testapp/goro.yaml uses.
		base := readTestdata(t, "add/goro.yaml")
		src := strings.Replace(base, workDirLine, `  work_dir: "" # the project, relative to this file`+"\n", 1)
		source, project := outsideProject(t, src)

		if err := commands.Init(readWithWorkDir(t, source, project), commands.Options{SkipGoMod: true, Out: io.Discard}); err != nil {
			t.Fatalf("init: %v", err)
		}

		if !exists(project, "goro.yaml") {
			t.Fatalf("init didn't save goro.yaml in the project root")
		}
		assertGolden(t, filepath.Join(project, "goro.yaml"), "add/goro.yaml")
	})
}

// TestInitWithGoroYamlInProjectRoot: init from a config outside the project never overwrites a
// goro.yaml the project root already holds (ADR 0001, ADR 0002).
func TestInitWithGoroYamlInProjectRoot(t *testing.T) {
	t.Run("a different goro.yaml makes init refuse", func(t *testing.T) {
		source, project := outsideProject(t, readTestdata(t, "init/goro.yaml"))
		writeFile(t, project, "goro.yaml", "app:\n  name: stale\n  module: stale\n  work_dir: .\n")
		cfg := readWithWorkDir(t, source, project)
		ageFiles(t, filepath.Dir(project))
		before := snapshot(t, filepath.Dir(project))

		err := commands.Init(cfg, commands.Options{SkipGoMod: true, Out: io.Discard})
		// The error names the project's goro.yaml, not the source next to it.
		assertError(t, err, errorCheck{file: filepath.Join(filepath.Base(project), "goro.yaml")})
		assertNoChanges(t, before, snapshot(t, filepath.Dir(project)))
	})

	t.Run("the goro.yaml init would write lets init go on", func(t *testing.T) {
		// Passes today, since init doesn't write goro.yaml yet; it keeps init from refusing the
		// copy it would write itself.
		source, project := outsideProject(t, readTestdata(t, "init/goro.yaml"))
		writeFile(t, project, "goro.yaml", readTestdata(t, "add/goro.yaml"))
		cfg := readWithWorkDir(t, source, project)

		if err := commands.Init(cfg, commands.Options{SkipGoMod: true, Out: io.Discard}); err != nil {
			t.Fatalf("init: %v", err)
		}
		assertGolden(t, filepath.Join(project, "goro.yaml"), "add/goro.yaml")
		if !exists(project, "internal/app/app.go") {
			t.Errorf("init didn't generate the project")
		}
	})
}
