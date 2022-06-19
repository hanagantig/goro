package chains

import (
	"github.com/spf13/afero"
	entity "goro/internal/config"
	"os"
)

type saveFilesChain struct {
	data entity.Config
}

func NewSaveFilesChain(data entity.Config) *saveFilesChain {
	return &saveFilesChain{
		data: data,
	}
}

func (g *saveFilesChain) Apply(fs *afero.Afero) (*afero.Afero, error) {
	err := fs.Walk(g.data.App.WorkDir,
		func(path string, f os.FileInfo, err error) error {
			if f.IsDir() {
				return os.MkdirAll(path, os.FileMode(0775))
			}

			file, err := os.Open(path)
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
