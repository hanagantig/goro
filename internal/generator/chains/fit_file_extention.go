package chains

import (
	"github.com/spf13/afero"
	entity "goro/internal/config"
	"os"
	"path/filepath"
	"strings"
)

type fitFileExtensionChain struct {
	data entity.Config
}

func NewFitFileExtensionChain(data entity.Config) *fitFileExtensionChain {
	return &fitFileExtensionChain{
		data: data,
	}
}

func (f *fitFileExtensionChain) Apply(fs *afero.Afero) (*afero.Afero, error) {
	err := filepath.WalkDir(f.data.App.WorkDir,
		func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			if strings.Contains(d.Name(), ".tmpl") {
				if _, err = os.Stat(path); !os.IsNotExist(err) {
					err = fs.Rename(path, strings.Replace(path, ".tmpl", "", 1))
					if err != nil {
						return err
					}
				}
			}

			return err
		},
	)

	return fs, err
}

func (f *fitFileExtensionChain) Name() string {
	return "Fit file extensions"
}

func (f *fitFileExtensionChain) Rollback() error {
	return nil
}
