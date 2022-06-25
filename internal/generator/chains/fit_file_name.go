package chains

import (
	entity "github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/pkg/afero"
	"github.com/iancoleman/strcase"
	"os"
	"strings"
)

type fitFileNameChain struct{}

func NewFitFileNameChain() *fitFileNameChain {
	return &fitFileNameChain{}
}

func (f *fitFileNameChain) Apply(fs *afero.Afero, data entity.Config) (*afero.Afero, error) {
	appName := strcase.ToKebab(data.App.Name)
	toRename := make([]string, 0, 0)

	err := fs.Walk("/", func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !f.IsDir() {
			return nil
		}

		if strings.Contains(f.Name(), "{{app_name}}") {
			if _, err = fs.Stat(path); !os.IsNotExist(err) {
				toRename = append(toRename, path)
			}
		}

		return nil
	})

	for _, p := range toRename {
		err = fs.Rename(p, strings.Replace(p, "{{app_name}}", appName, 1))
		if err != nil {
			return nil, err
		}
		fs.RemoveAll(p)
	}

	return fs, err
}

func (f *fitFileNameChain) Name() string {
	return "Fit file names"
}

func (f *fitFileNameChain) Rollback() error {
	return nil
}
