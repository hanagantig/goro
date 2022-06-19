package chains

import (
	"errors"
	"github.com/spf13/afero"
	entity "goro/internal/config"
	"os/exec"
)

type modInitChain struct {
	data entity.Config
}

func NewModInitChain(data entity.Config) *modInitChain {
	return &modInitChain{
		data: data,
	}
}

func (m *modInitChain) Apply(fs *afero.Afero) (*afero.Afero, error) {
	cmd := exec.Command("go", "mod", "init", m.data.App.Name)
	cmd.Dir = m.data.App.WorkDir

	output, err := cmd.CombinedOutput()

	out := string(output)
	if err != nil && out != "" {
		return fs, errors.New(out)
	}

	return fs, nil
}

func (g *modInitChain) Name() string {
	return "Go mod init"
}

func (g *modInitChain) Rollback() error {
	return nil
}
