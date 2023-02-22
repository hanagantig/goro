package chains

import (
	entity "github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/pkg/afero"
	"os"
	"strings"
)

type generateServicesChain struct{}

func NewGenerateServicesChain() *generateServicesChain {
	return &generateServicesChain{}
}

func (g *generateServicesChain) Apply(fs *afero.Afero, data entity.Config) (*afero.Afero, error) {
	servicePath := "/internal/service"
	serviceFilePath := "/internal/service/service.go.tmpl"

	svcTmpl, err := fs.ReadFile(serviceFilePath)
	if err != nil {
		return nil, err
	}

	err = fs.Remove(serviceFilePath)
	if err != nil {
		return nil, err
	}

	for svcName, _ := range data.Services {
		pkgName := strings.ToLower(svcName.String())
		path := servicePath + "/" + pkgName
		if _, err := fs.Stat(path); os.IsNotExist(err) {
			err = fs.Mkdir(path, os.ModeDir)
			if err != nil {
				return nil, err
			}
		}

		generated, err := generate(path, svcTmpl, data)
		if err != nil {
			return nil, err
		}

		err = fs.WriteFile(path+"/service.go.tmpl", generated, 0644)
		if err != nil {
			return nil, err
		}
	}

	return fs, nil
}

func (g *generateServicesChain) Name() string {
	return "Generate services chain"
}

func (g *generateServicesChain) Rollback() error {
	return nil
}
