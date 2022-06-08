package chains

import (
	"errors"
	"os/exec"
)

type modTidyChain struct {
	workDir string
}

func NewModTidyChain(workDir string) *modTidyChain {
	return &modTidyChain{
		workDir: workDir,
	}
}

func (m *modTidyChain) Apply() error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = m.workDir

	output, err := cmd.CombinedOutput()
	out := string(output)
	if err != nil && out != "" {
		return errors.New(out)
	}

	return nil
}

func (g *modTidyChain) Name() string {
	return "Go mod tidy"
}

func (g *modTidyChain) Rollback() error {
	return nil
}
