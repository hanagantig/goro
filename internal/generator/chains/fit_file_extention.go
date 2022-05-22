package chains

import (
	entity "goro/internal/entity"
	"os"
	"path/filepath"
	"strings"
)

type fitFileExtensionChain struct {
	data entity.AppData
}

func NewFitFileExtensionChain(data entity.AppData) *fitFileExtensionChain {
	return &fitFileExtensionChain{
		data: data,
	}
}

func (f *fitFileExtensionChain) Apply() error {
	err := filepath.WalkDir("/Users/hanagantig/tmp/gorotest",
		func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			if strings.Contains(d.Name(), ".tmpl") {
				if _, err = os.Stat(path); !os.IsNotExist(err) {
					err = os.Rename(path, strings.Replace(path, ".tmpl", "", 1))
					if err != nil {
						return err
					}
				}
			}

			return err
		},
	)

	return err
}

func (f *fitFileExtensionChain) Name() string {
	return "Fit file extensions"
}

func (f *fitFileExtensionChain) Rollback() error {
	return nil
}
