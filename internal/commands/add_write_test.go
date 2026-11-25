package commands_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/hanagantig/goro/internal/commands"
)

// TestAddFailingAfterGoroYamlIsWritten: add writes goro.yaml first, so an error after that, such
// as a file it can't write, says that goro.yaml lists the addition now and that goro update
// finishes the job once the cause is fixed.
func TestAddFailingAfterGoroYamlIsWritten(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs a directory the test can't write to")
	}

	dir, path := addProject(t)
	usecase := filepath.Join(dir, "internal", "usecase")
	if err := os.Chmod(usecase, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(usecase, 0o775) })

	_, err := commands.AddUseCase(path, "SignOut", addOpts)
	if err == nil {
		t.Fatalf("add usecase SignOut into a read-only internal/usecase succeeded")
	}
	for _, want := range []string{"now lists", "SignOut", "goro update"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q doesn't say %q", err, want)
		}
	}
	if !strings.Contains(readFile(t, dir, "goro.yaml"), "- SignOut") {
		t.Errorf("goro.yaml doesn't list SignOut after the failed add")
	}

	// The fix the error asks for works.
	if err := os.Chmod(usecase, 0o775); err != nil {
		t.Fatal(err)
	}
	cfg, err := commands.LoadConfig(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	report := update(t, cfg)
	assertCreated(t, report, dir, "internal/usecase/sign_out.go")
}

// TestAddWritesGoroYamlInPlace: add replaces goro.yaml whole (through a temporary file and a
// rename), keeps its mode, writes through a symlink to its target and leaves no other file.
func TestAddWritesGoroYamlInPlace(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file modes and symlinks differ on Windows")
	}

	t.Run("the mode stays", func(t *testing.T) {
		dir, path := addProject(t)
		if err := os.Chmod(path, 0o600); err != nil {
			t.Fatal(err)
		}
		before, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := commands.AddUseCase(path, "SignOut", addOpts); err != nil {
			t.Fatalf("add usecase SignOut: %v", err)
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Errorf("goro.yaml has mode %v after add, want %v", info.Mode().Perm(), os.FileMode(0o600))
		}
		after, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(after) != len(before) {
			t.Errorf("the project root holds %d entries after add, %d before: %v", len(after), len(before), after)
		}
	})

	t.Run("a symlink stays", func(t *testing.T) {
		_, path := addProject(t)
		shared := filepath.Join(t.TempDir(), "goro.yaml")
		if err := os.Rename(path, shared); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(shared, path); err != nil {
			t.Fatal(err)
		}

		if _, err := commands.AddUseCase(path, "SignOut", addOpts); err != nil {
			t.Fatalf("add usecase SignOut: %v", err)
		}

		info, err := os.Lstat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("goro.yaml is no longer a symlink after add")
		}
		if !strings.Contains(readFile(t, filepath.Dir(shared), "goro.yaml"), "- SignOut") {
			t.Errorf("the symlink's target doesn't list SignOut")
		}
	})
}
