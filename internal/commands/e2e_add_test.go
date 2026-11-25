package commands_test

import (
	"io"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hanagantig/goro/internal/commands"
)

// TestE2EInitAddBuild runs the full pipeline, go mod init and tidy included, the way a user grows
// a project: init from a goro.yaml outside the project, then every goro add subcommand against
// the goro.yaml init saved in the project root, then a build of the result. It needs the network.
func TestE2EInitAddBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e: runs go mod tidy and go build, which need the network")
	}

	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("the e2e test needs the go tool: %v", err)
	}

	// testdata/init/goro.yaml lives in root/templates and has work_dir: ../app.
	root := t.TempDir()
	project := filepath.Join(root, "app")
	mkdir(t, project)
	writeFile(t, root, "templates/goro.yaml", readTestdata(t, "init/goro.yaml"))

	cfg, err := commands.LoadConfig(filepath.Join(root, "templates", "goro.yaml"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	// Init writes into the work dir, so stop before it unless work_dir: ../app resolved to the
	// project.
	if !sameDir(cfg.App.WorkDir, project) {
		t.Fatalf("LoadConfig gives work dir %q for work_dir: ../app, want %s", cfg.App.WorkDir, project)
	}
	if err := commands.Init(cfg, commands.Options{Out: io.Discard}); err != nil {
		t.Fatalf("init: %v", err)
	}
	if !exists(project, "goro.yaml") {
		t.Fatalf("init didn't save goro.yaml in the project root")
	}

	path := filepath.Join(project, "goro.yaml")
	runAddSteps(t, path, commands.Options{Out: io.Discard})
	assertGolden(t, path, "add/goro.golden.yaml")

	build := exec.Command(goBin, "build", "./...")
	build.Dir = project
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./... in the generated app: %v\n%s", err, out)
	}
}
