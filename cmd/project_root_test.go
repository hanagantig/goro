package cmd_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hanagantig/goro/internal/commands"
	"github.com/hanagantig/goro/internal/config"
)

// TestNothingToAskAfterInitInTheProjectRoot: when init runs on the goro.yaml in the project root
// and has to ask for the app name and the work dir, it writes the answers into that goro.yaml
// (ADR 0002), so update and add, run in the project root, have nothing to ask. Their stdin is
// empty here, so a question would fail.
func TestNothingToAskAfterInitInTheProjectRoot(t *testing.T) {
	t.Chdir(t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "goro.yaml")
	yaml := strings.Replace(projectConfig, "  name: demo\n", "", 1)
	yaml = strings.Replace(yaml, "  work_dir: .\n", "  work_dir: \"\"\n", 1)
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.NewConfig(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	// The answers to init's questions.
	cfg.App.Name = "demo"
	cfg.App.WorkDir = dir
	if err := commands.Init(cfg, commands.Options{SkipGoMod: true, Out: io.Discard}); err != nil {
		t.Fatalf("init: %v", err)
	}

	for _, args := range [][]string{
		{"update", "--dry-run"},
		{"add", "usecase", "SignOut", "--dry-run"},
	} {
		out, err := goro(t, dir, args...)
		if err != nil {
			t.Errorf("goro %s in the project root after init: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
}

// TestNoGoroYamlInTheCurrentDir: init and update in a directory without goro.yaml say how to
// find one: run in the project root or pass --config.
func TestNoGoroYamlInTheCurrentDir(t *testing.T) {
	t.Chdir(t.TempDir())
	dir := t.TempDir()

	for _, args := range [][]string{{"init"}, {"update"}, {"update", "--dry-run"}} {
		out, err := goro(t, dir, args...)
		if err == nil {
			t.Errorf("goro %s without goro.yaml succeeded:\n%s", strings.Join(args, " "), out)
		}
		if !strings.Contains(out, "--config") || !strings.Contains(out, "project root") {
			t.Errorf("goro %s without goro.yaml doesn't point to the project root or --config:\n%s", strings.Join(args, " "), out)
		}
	}
}
