package chains

import (
	entity "github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/pkg/afero"
	"os"
	"path/filepath"
	"strings"
)

type syncAdaptersChain struct {
	data entity.Config
}

func NewSyncAdaptersChain() *syncAdaptersChain {
	return &syncAdaptersChain{}
}

func (g *syncAdaptersChain) Apply(fs *afero.Afero, data entity.Config) (*afero.Afero, error) {

	adapterPath := filepath.Join(data.App.WorkDir, "/internal/adapter")

	var loadedAdapters []string
	for _, adapter := range data.Adapters {
		typeFolder := adapter.Storage.String() + "repo"
		loadedAdapters = append(loadedAdapters, strings.ToLower(filepath.Join(typeFolder, adapter.Name)))
	}

	typeFolders, err := os.ReadDir(adapterPath)
	if err != nil {
		return fs, err
	}

	for _, types := range typeFolders {
		if types.IsDir() {
			adaptersFolder, err := os.ReadDir(filepath.Join(adapterPath, types.Name()))
			for _, af := range adaptersFolder {
				if !af.IsDir() {
					continue
				}
				if !contains(loadedAdapters, strings.ToLower(filepath.Join(types.Name(), af.Name()))) {
					err = os.RemoveAll(filepath.Join(adapterPath, types.Name(), af.Name()))
					if err != nil {
						return fs, err
					}
				}
			}
		}
	}

	return fs, nil
}

func (g *syncAdaptersChain) Name() string {
	return "Synchronize adapters"
}

func (g *syncAdaptersChain) Rollback() error {
	return nil
}
