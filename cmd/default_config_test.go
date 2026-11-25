package cmd_test

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hanagantig/goro/cmd"
	"github.com/hanagantig/goro/internal/commands"
)

// Rules under test: ADR 0002, update and add read ./goro.yaml when --config isn't given, and
// --config points them elsewhere. These tests run the command line; the commands package tests
// everything else.

// runMainEnv makes the test binary run goro's command line instead of the tests, so a test can
// run goro in another directory without building it first.
const runMainEnv = "GORO_CMD_TEST_RUN_MAIN"

func TestMain(m *testing.M) {
	if os.Getenv(runMainEnv) == "1" {
		cmd.Execute()
		os.Exit(0)
	}

	os.Exit(m.Run())
}

// projectConfig is the goro.yaml in the root of the test project.
const projectConfig = `app:
  name: demo
  module: demo
  work_dir: .
storages:
  - mysql
use_case:
  deps:
    - Users
  methods:
    - SignIn
services:
  - name: Users
    methods:
      - Create
    deps:
      - UserRepo
adapters:
  - name: UserRepo
    storage: mysql
    methods:
      - GetByID
`

// newProject makes a project with projectConfig in its root and returns its dir. The test moves
// into an empty working directory first, so nothing resolved against it lands in goro's tree.
func newProject(t *testing.T) string {
	t.Helper()

	t.Chdir(t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "goro.yaml")
	if err := os.WriteFile(path, []byte(projectConfig), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := commands.LoadConfig(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	// The setup sets the work dir itself, so it doesn't depend on how work_dir: . resolves.
	cfg.App.WorkDir = dir
	if err := commands.Init(cfg, commands.Options{SkipGoMod: true, Out: io.Discard}); err != nil {
		t.Fatalf("init: %v", err)
	}

	return dir
}

// goro runs goro's command line with args in dir, as a user would in a shell there, and returns
// its combined output. Stdin is empty, so a question fails instead of waiting for an answer.
func goro(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()

	bin, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c := exec.CommandContext(ctx, bin, args...)
	c.Dir = dir
	c.Env = append(os.Environ(), runMainEnv+"=1")
	out, err := c.CombinedOutput()

	return string(out), err
}

type fileState struct {
	content string
	modTime time.Time
}

// files records every file under dir by its slash-separated relative path.
func files(t *testing.T, dir string) map[string]fileState {
	t.Helper()

	res := make(map[string]fileState)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		res[filepath.ToSlash(rel)] = fileState{content: string(b), modTime: info.ModTime()}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	return res
}

func assertSameFiles(t *testing.T, before, after map[string]fileState) {
	t.Helper()

	for path, b := range before {
		a, ok := after[path]
		switch {
		case !ok:
			t.Errorf("%s was deleted", path)
		case a != b:
			t.Errorf("%s was written", path)
		}
	}
	for path := range after {
		if _, ok := before[path]; !ok {
			t.Errorf("%s was created", path)
		}
	}
}

func hasLineWith(text string, parts ...string) bool {
	for _, line := range strings.Split(text, "\n") {
		all := true
		for _, p := range parts {
			if !strings.Contains(line, p) {
				all = false
				break
			}
		}
		if all {
			return true
		}
	}

	return false
}

func TestUpdateReadsGoroYamlInCurrentDir(t *testing.T) {
	dir := newProject(t)

	// goro.yaml: a new use case method.
	path := filepath.Join(dir, "goro.yaml")
	edited := strings.Replace(projectConfig, "    - SignIn\n", "    - SignIn\n    - SignOut\n", 1)
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	before := files(t, dir)

	out, err := goro(t, dir, "update", "--dry-run")
	if err != nil {
		t.Fatalf("goro update --dry-run in the project root, without --config: %v\n%s", err, out)
	}
	if !hasLineWith(out, "create", "internal/usecase/sign_out.go") {
		t.Errorf("goro update --dry-run doesn't show the new use case method from ./goro.yaml:\n%s", out)
	}
	assertSameFiles(t, before, files(t, dir))
}

func TestAddReadsGoroYamlInCurrentDir(t *testing.T) {
	cases := []struct {
		args []string
		// creates is a file the add would create.
		creates string
	}{
		{args: []string{"add", "usecase", "SignOut"}, creates: "internal/usecase/sign_out.go"},
		{args: []string{"add", "service", "Billing", "--deps", "UserRepo", "--methods", "Charge,Refund"}, creates: "internal/service/billing/refund.go"},
		{args: []string{"add", "method", "Users.Delete"}, creates: "internal/service/users/delete.go"},
		{args: []string{"add", "adapter", "PaymentsRepo", "--storage", "pgsqlx", "--methods", "Save"}, creates: "internal/adapter/pgsqlxrepo/paymentsrepo/save.go"},
	}

	for _, c := range cases {
		t.Run(strings.Join(c.args, " "), func(t *testing.T) {
			dir := newProject(t)
			before := files(t, dir)

			// --dry-run, so add leaves out go mod tidy and writes nothing.
			args := append(append([]string{}, c.args...), "--dry-run")
			out, err := goro(t, dir, args...)
			if err != nil {
				t.Fatalf("goro %s in the project root, without --config: %v\n%s", strings.Join(args, " "), err, out)
			}
			if !hasLineWith(out, "create", c.creates) {
				t.Errorf("goro %s doesn't show that it would create %s:\n%s", strings.Join(args, " "), c.creates, out)
			}
			assertSameFiles(t, before, files(t, dir))
		})
	}
}

func TestConfigFlagOverridesTheDefault(t *testing.T) {
	project := newProject(t)
	path := filepath.Join(project, "goro.yaml")
	// A directory without goro.yaml to run goro in.
	elsewhere := t.TempDir()

	t.Run("update", func(t *testing.T) {
		// goro.yaml: a new use case method.
		edited := strings.Replace(projectConfig, "    - SignIn\n", "    - SignIn\n    - SignOut\n", 1)
		if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.WriteFile(path, []byte(projectConfig), 0o644); err != nil {
				t.Fatal(err)
			}
		})
		before := files(t, project)

		out, err := goro(t, elsewhere, "update", "--dry-run", "--config", path)
		if err != nil {
			t.Fatalf("goro update --dry-run --config %s: %v\n%s", path, err, out)
		}
		if !hasLineWith(out, "create", "internal/usecase/sign_out.go") {
			t.Errorf("goro update --dry-run --config doesn't show the new use case method:\n%s", out)
		}
		// The project's work_dir is ., so the dry run has to look at the project, where
		// app.go exists, not at the directory goro runs in.
		if hasLineWith(out, "create", "internal/app/app.go") {
			t.Errorf("goro update --dry-run --config planned the project in the directory it runs in:\n%s", out)
		}
		assertSameFiles(t, before, files(t, project))
	})

	t.Run("add", func(t *testing.T) {
		before := files(t, project)

		out, err := goro(t, elsewhere, "add", "usecase", "SignOut", "--dry-run", "--config", path)
		if err != nil {
			t.Fatalf("goro add usecase SignOut --dry-run --config %s: %v\n%s", path, err, out)
		}
		if !hasLineWith(out, "create", "internal/usecase/sign_out.go") {
			t.Errorf("goro add usecase --dry-run --config doesn't show that it would create sign_out.go:\n%s", out)
		}
		if hasLineWith(out, "create", "internal/app/app.go") {
			t.Errorf("goro add --dry-run --config planned the project in the directory it runs in:\n%s", out)
		}
		assertSameFiles(t, before, files(t, project))
	})

	if entries, err := os.ReadDir(elsewhere); err != nil || len(entries) != 0 {
		t.Errorf("goro wrote into the directory it ran in: %v %v", entries, err)
	}
}
