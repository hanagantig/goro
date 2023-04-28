package chains

import (
	entity "github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/pkg/afero"
	"os"
	"path/filepath"
	"strings"
)

type updateFilesChain struct {
	data entity.Config
}

func NewUpdateFilesChain() *updateFilesChain {
	return &updateFilesChain{}
}

func isRegenerated(name string) bool {
	return strings.Contains(name, "internal/app/container.go")
}

func (g *updateFilesChain) Apply(fs *afero.Afero, data entity.Config) (*afero.Afero, error) {
	err := fs.Walk("/",
		func(path string, f os.FileInfo, err error) error {
			appPath := filepath.Join(data.App.WorkDir, path)

			_, err = os.Stat(appPath)
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			if err == nil && !isRegenerated(appPath) {
				return nil
			}

			if f.IsDir() {
				return os.MkdirAll(appPath, os.FileMode(0775))
			}

			file, err := os.Create(appPath)
			if err != nil {
				return err
			}

			content, err := fs.ReadFile(path)
			if err != nil {
				return err
			}

			_, err = file.Write(content)
			if err != nil {
				return err
			}

			return nil
		},
	)

	return fs, err
}

func (g *updateFilesChain) Name() string {
	return "Updates file in workdir"
}

func (g *updateFilesChain) Rollback() error {
	return nil
}
