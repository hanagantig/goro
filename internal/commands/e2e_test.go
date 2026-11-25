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
// config: init, hand edits, a goro.yaml change, update, then a build of the result. It needs the
// network.
func TestE2EInitUpdateBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e: runs go mod tidy and go build, which need the network")
	}

	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("the e2e test needs the go tool: %v", err)
	}

	dir := t.TempDir()

	// Init from the example config without the pgsqlx storage and its UserRepo adapter.
	initCfg := exampleConfig(t, dir)
	storages := make(config.Storages, 0, len(initCfg.Storages))
	for _, s := range initCfg.Storages {
		if s != "pgsqlx" {
			storages = append(storages, s)
		}
	}
	initCfg.Storages = storages
	adapters := make(config.Adapters, 0, len(initCfg.Adapters))
	for _, a := range initCfg.Adapters {
		if a.Name != "UserRepo" {
			adapters = append(adapters, a)
		}
	}
	initCfg.Adapters = adapters

	if err := commands.Init(initCfg, commands.Options{}); err != nil {
		t.Fatalf("init: %v", err)
	}

	// Hand edits in a method file and a dep file, and a hand-made package.
	kept := make(map[string]string)
	for path, fn := range map[string]string{
		"internal/usecase/sign_in.go":               "handWrittenSignIn",
		"internal/service/myservice/dep_my_repo.go": "handWrittenMyRepoDep",
	} {
		if !exists(dir, path) {
			t.Errorf("init didn't create %s", path)
			continue
		}
		kept[path] = handEditFunc(readFile(t, dir, path), fn)
		writeFile(t, dir, path, kept[path])
	}
	kept["internal/service/mailer/mailer.go"] = "package mailer\n\nfunc Send() error { return nil }\n"
	writeFile(t, dir, "internal/service/mailer/mailer.go", kept["internal/service/mailer/mailer.go"])

	// goro.yaml: pgsqlx and UserRepo back, OrderService also depends on MyRepo, PingPong removed.
	cfg := exampleConfig(t, dir)
	services := make(config.Services, 0, len(cfg.Services))
	for _, s := range cfg.Services {
		if s.Name == "PingPong" {
			continue
		}
		if s.Name == "OrderService" {
			s.Deps = append(s.Deps, "MyRepo")
		}
		services = append(services, s)
	}
	cfg.Services = services
	ucDeps := make([]string, 0, len(cfg.UseCase.Deps))
	for _, d := range cfg.UseCase.Deps {
		if d != "PingPong" {
			ucDeps = append(ucDeps, d)
		}
	}
	cfg.UseCase.Deps = ucDeps

	if _, err := commands.Update(cfg, commands.Options{Out: io.Discard}); err != nil {
		t.Fatalf("update: %v", err)
	}

	for path, content := range kept {
		if !exists(dir, path) {
			t.Errorf("update deleted %s", path)
			continue
		}
		if readFile(t, dir, path) != content {
			t.Errorf("update lost the hand edits in %s", path)
		}
	}

	build := exec.Command(goBin, "build", "./...")
	build.Dir = dir
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./... in the generated app: %v\n%s", err, out)
	}
}

// exampleConfig loads example/testapp/goro.yaml and points it at workDir.
func exampleConfig(t *testing.T, workDir string) config.Config {
	t.Helper()

	cfg, err := config.NewConfig(filepath.Join("..", "..", "example", "testapp", "goro.yaml"))
	if err != nil {
		t.Fatalf("load example config: %v", err)
	}
	cfg.App.WorkDir = workDir

	return cfg
}
