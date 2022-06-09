package chains

import (
	"errors"
	entity "goro/internal/entity"
	"os/exec"
)

type modInitChain struct {
	data entity.AppData
}

func NewModInitChain(data entity.AppData) *modInitChain {
	return &modInitChain{
		data: data,
	}
}

func (m *modInitChain) Apply() error {
	cmd := exec.Command("go", "mod", "init", m.data.App.Name)
	cmd.Dir = m.data.App.WorkDir

	output, err := cmd.CombinedOutput()

	out := string(output)
	if err != nil && out != "" {
		return errors.New(out)
	}

	return nil
}

func (g *modInitChain) Name() string {
	return "Go mod init"
}

func (g *modInitChain) Rollback() error {
	return nil
}
