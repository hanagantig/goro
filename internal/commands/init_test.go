package commands_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/hanagantig/goro/internal/commands"
)

func TestInitRefusesExistingProject(t *testing.T) {
	dir := t.TempDir()
	initProject(t, newConfig(t, dir))

	const path = "internal/usecase/sign_in.go"
	edited := handEdit(readFile(t, dir, path))
	writeFile(t, dir, path, edited)

	err := commands.Init(newConfig(t, dir), commands.Options{SkipGoMod: true})
	switch {
	case err == nil:
		t.Errorf("init succeeded in a work dir that already holds a goro project")
	case !errors.Is(err, commands.ErrProjectExists):
		t.Errorf("init error = %v, want ErrProjectExists", err)
	case !strings.Contains(err.Error(), "update"):
		t.Errorf("init error %q doesn't point to goro update", err)
	}

	if readFile(t, dir, path) != edited {
		t.Errorf("init overwrote %s in an existing project", path)
	}
}
