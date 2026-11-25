package commands_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hanagantig/goro/internal/commands"
)

// TestSummaryNamesTheAbsoluteWorkDir: the summary line names the work dir as an absolute path,
// also when goro.yaml gives a relative one.
func TestSummaryNamesTheAbsoluteWorkDir(t *testing.T) {
	t.Chdir(t.TempDir())
	writeFile(t, ".", "app/goro.yaml", readTestdata(t, "add/goro.yaml"))
	abs, err := filepath.Abs("app")
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := commands.LoadConfig(filepath.Join("app", "goro.yaml"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	var out bytes.Buffer
	if err := commands.Init(cfg, commands.Options{SkipGoMod: true, Out: &out}); err != nil {
		t.Fatalf("init: %v", err)
	}
	if !strings.Contains(out.String(), "goro init in "+abs+":") {
		t.Errorf("init's summary doesn't name the work dir %s:\n%s", abs, out.String())
	}

	out.Reset()
	if _, err := commands.AddUseCase(filepath.Join("app", "goro.yaml"), "SignOut", commands.Options{SkipGoMod: true, Out: &out}); err != nil {
		t.Fatalf("add usecase SignOut: %v", err)
	}
	if !strings.Contains(out.String(), "goro add in "+abs+":") {
		t.Errorf("add's summary doesn't name the work dir %s:\n%s", abs, out.String())
	}
}

// TestAddDryRunNamesGoroYamlAsGiven: the dry run of add names goro.yaml by the path it was given,
// cleaned, in its action line and its diff header, with git's a/ and b/ only for a path below the
// current directory.
func TestAddDryRunNamesGoroYamlAsGiven(t *testing.T) {
	dir, _ := addProject(t)
	t.Chdir(dir)

	cases := []struct {
		name, given, shown, from, to string
	}{
		{name: "a relative path", given: "./goro.yaml", shown: "goro.yaml", from: "a/goro.yaml", to: "b/goro.yaml"},
		{
			name:  "an absolute path",
			given: dir + string(filepath.Separator) + "." + string(filepath.Separator) + "goro.yaml",
			shown: filepath.ToSlash(filepath.Join(dir, "goro.yaml")),
			from:  filepath.ToSlash(filepath.Join(dir, "goro.yaml")),
			to:    filepath.ToSlash(filepath.Join(dir, "goro.yaml")),
		},
		{
			name:  "a path up from the current directory",
			given: filepath.Join("..", filepath.Base(dir), "goro.yaml"),
			shown: "../" + filepath.Base(dir) + "/goro.yaml",
			from:  "../" + filepath.Base(dir) + "/goro.yaml",
			to:    "../" + filepath.Base(dir) + "/goro.yaml",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out bytes.Buffer
			if _, err := commands.AddUseCase(c.given, "SignOut", commands.Options{DryRun: true, Out: &out}); err != nil {
				t.Fatalf("add usecase SignOut --dry-run: %v", err)
			}
			printed := out.String()
			if !hasLineWith(printed, "overwrite", " "+c.shown) {
				t.Errorf("no overwrite line for %s:\n%s", c.shown, printed)
			}
			if !strings.Contains(printed, "--- "+c.from+"\n+++ "+c.to+"\n") {
				t.Errorf("no diff header from %s to %s:\n%s", c.from, c.to, printed)
			}
		})
	}
}
