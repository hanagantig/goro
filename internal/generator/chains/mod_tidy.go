package chains

import (
	"errors"
	"github.com/spf13/afero"
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

func (m *modTidyChain) Apply(fs *afero.Afero) (*afero.Afero, error) {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = m.workDir

	output, err := cmd.CombinedOutput()
	out := string(output)
	if err != nil && out != "" {
		return fs, errors.New(out)
	}

	return fs, nil
}

func (m *modTidyChain) Name() string {
	return "Go mod tidy"
}

func (m *modTidyChain) Rollback() error {
	return nil
}
