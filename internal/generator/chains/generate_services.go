package chains

import (
	entity "github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/pkg/afero"
	"github.com/iancoleman/strcase"
	"os"
)

type generateServicesChain struct{}

func NewGenerateServicesChain() *generateServicesChain {
	return &generateServicesChain{}
}

func (g *generateServicesChain) Apply(fs *afero.Afero, data entity.Config) (*afero.Afero, error) {
	servicePath := "/internal/service"
	serviceFilePath := "/internal/service/service.go.tmpl"
	methodFilePath := "/internal/service/method.go.tmpl"

	svcTmpl, err := fs.ReadFile(serviceFilePath)
	if err != nil {
		return nil, err
	}

	methodTmpl, err := fs.ReadFile(methodFilePath)
	if err != nil {
		return nil, err
	}

	err = fs.Remove(serviceFilePath)
	if err != nil {
		return nil, err
	}

	err = fs.Remove(methodFilePath)
	if err != nil {
		return nil, err
	}

	for _, svc := range data.Services {
		path := servicePath + "/" + svc.GetPkgName()
		if _, err := fs.Stat(path); os.IsNotExist(err) {
			err = fs.Mkdir(path, os.ModeDir)
			if err != nil {
				return nil, err
			}
		}

		generated, err := generate(path, svcTmpl, svc)
		if err != nil {
			return nil, err
		}

		err = fs.WriteFile(path+"/service.go.tmpl", generated, 0644)
		if err != nil {
			return nil, err
		}

		for _, method := range svc.Methods {
			methodData := struct {
				PkgName    string
				MethodName string
			}{
				PkgName:    svc.GetPkgName(),
				MethodName: method,
			}
			generatedMethod, err := generate(path, methodTmpl, methodData)
			if err != nil {
				return nil, err
			}

			err = fs.WriteFile(path+"/"+strcase.ToSnake(method)+".go", generatedMethod, 0644)
			if err != nil {
				return nil, err
			}
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
