package chains

import (
	"github.com/iancoleman/strcase"
	"github.com/spf13/afero"
	entity "goro/internal/config"
	"os"
	"strings"
)

type fitFileNameChain struct {
	data entity.Config
}

func NewFitFileNameChain(data entity.Config) *fitFileNameChain {
	return &fitFileNameChain{
		data: data,
	}
}

func (f *fitFileNameChain) Apply(fs *afero.Afero) (*afero.Afero, error) {
	appName := strcase.ToKebab(f.data.App.Name)
	toRename := make([]string, 0, 0)

	err := fs.Walk("templates/app", func(path string, f os.FileInfo, err error) error {
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
	}

	return fs, err
}

func (f *fitFileNameChain) Name() string {
	return "Fit file names"
}

func (f *fitFileNameChain) Rollback() error {
	return nil
}
