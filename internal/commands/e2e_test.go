package commands_test

import (
	"io"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hanagantig/goro/internal/commands"
	"github.com/hanagantig/goro/internal/config"
)

// TestE2EInitUpdateBuild runs the full pipeline, go mod init and tidy included, on the example
// config and builds the result. It needs the network.
func TestE2EInitUpdateBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e: runs go mod tidy and go build, which need the network")
	}

	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("the e2e test needs the go tool: %v", err)
	}

	cfg, err := config.NewConfig(filepath.Join("..", "..", "example", "testapp", "goro.yaml"))
	if err != nil {
		t.Fatalf("load example config: %v", err)
	}
	dir := t.TempDir()
	cfg.App.WorkDir = dir

	if err := commands.Init(cfg, commands.Options{}); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := commands.Update(cfg, commands.Options{Out: io.Discard}); err != nil {
		t.Fatalf("update: %v", err)
	}

	build := exec.Command(goBin, "build", "./...")
	build.Dir = dir
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./... in the generated app: %v\n%s", err, out)
	}
}
