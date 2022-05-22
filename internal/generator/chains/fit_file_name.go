package chains

import (
	"fmt"
	"github.com/iancoleman/strcase"
	entity "goro/internal/entity"
	"os"
	"path/filepath"
	"strings"
)

type fitFileNameChain struct {
	data entity.AppData
}

func NewFitFileNameChain(data entity.AppData) *fitFileNameChain {
	return &fitFileNameChain{
		data: data,
	}
}

func (f *fitFileNameChain) Apply() error {
	appName := strcase.ToKebab(f.data.Name)
	toRename := make([]string, 0, 0)
	err := filepath.WalkDir("/Users/hanagantig/tmp/gorotest",
		func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				return nil
			}

			if strings.Contains(d.Name(), "{{app_name}}") {
				if _, err = os.Stat(path); !os.IsNotExist(err) {
					toRename = append(toRename, path)
				}
			}

			return nil
		},
	)

	for _, p := range toRename {
		err = os.Rename(p, strings.Replace(p, "{{app_name}}", appName, 1))
		fmt.Println(err)
		if err != nil {
			return err
		}
	}

	return err
}

func (f *fitFileNameChain) Name() string {
	return "Fit file names"
}

func (f *fitFileNameChain) Rollback() error {
	return nil
}
