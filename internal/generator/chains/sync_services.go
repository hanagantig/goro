package chains

import (
	entity "github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/pkg/afero"
	"os"
	"path/filepath"
	"strings"
)

type syncServicesChain struct {
	data entity.Config
}

func NewSyncServicesChain() *syncServicesChain {
	return &syncServicesChain{}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func (g *syncServicesChain) Apply(fs *afero.Afero, data entity.Config) (*afero.Afero, error) {

	servicePath := filepath.Join(data.App.WorkDir, "/internal/service")

	var loadedServices []string
	for _, service := range data.Services {
		loadedServices = append(loadedServices, strings.ToLower(service.Name))
	}

	servicesFolder, err := os.ReadDir(servicePath)
	for _, sf := range servicesFolder {
		if sf.IsDir() {
			if !contains(loadedServices, strings.ToLower(sf.Name())) {
				err = os.RemoveAll(filepath.Join(servicePath, sf.Name()))
				if err != nil {
					return fs, err
				}
			}
		}
	}

	return fs, nil
}

func (g *syncServicesChain) Name() string {
	return "Synchronize services"
}

func (g *syncServicesChain) Rollback() error {
	return nil
}
