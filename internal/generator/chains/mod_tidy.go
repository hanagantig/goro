package chains

import (
	"errors"
	entity "github.com/hanagantig/goro/internal/config"
	"github.com/spf13/afero"
	"os/exec"
)

type modTidyChain struct{}

func NewModTidyChain() *modTidyChain {
	return &modTidyChain{}
}

func (m *modTidyChain) Apply(fs *afero.Afero, data entity.Config) (*afero.Afero, error) {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = data.App.WorkDir

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
