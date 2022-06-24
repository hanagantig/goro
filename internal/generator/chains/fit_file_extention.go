package chains

import (
	"github.com/spf13/afero"
	entity "goro/internal/config"
	"os"
	"strings"
)

type fitFileExtensionChain struct{}

func NewFitFileExtensionChain() *fitFileExtensionChain {
	return &fitFileExtensionChain{}
}

func (f *fitFileExtensionChain) Apply(fs *afero.Afero, data entity.Config) (*afero.Afero, error) {
	err := fs.Walk("/",
		func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if f.IsDir() {
				return nil
			}

			if strings.Contains(f.Name(), ".tmpl") {
				if _, err = fs.Stat(path); !os.IsNotExist(err) {
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
