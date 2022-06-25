package chains

import (
	entity "github.com/hanagantig/goro/internal/config"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
)

type saveFilesChain struct {
	data entity.Config
}

func NewSaveFilesChain() *saveFilesChain {
	return &saveFilesChain{}
}

func (g *saveFilesChain) Apply(fs *afero.Afero, data entity.Config) (*afero.Afero, error) {
	err := fs.Walk("/",
		func(path string, f os.FileInfo, err error) error {
			appPath := filepath.Join(data.App.WorkDir, path)
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

func (g *saveFilesChain) Name() string {
	return "Save files to work dir"
}

func (g *saveFilesChain) Rollback() error {
	return nil
}
