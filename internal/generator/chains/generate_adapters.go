package chains

import (
	"fmt"
	entity "github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/pkg/afero"
	"github.com/iancoleman/strcase"
	"os"
	"strings"
)

type generateAdapterChain struct{}

func NewGenerateAdapterChain() *generateAdapterChain {
	return &generateAdapterChain{}
}

func (g *generateAdapterChain) Apply(fs *afero.Afero, data entity.Config) (*afero.Afero, error) {
	adapterPath := "/internal/adapter"
	repoFilePath := adapterPath + "/repository.go.tmpl"
	methodFilePath := "/internal/adapter/method.go.tmpl"
	transactorFilePath := "/internal/adapter/sql_transactor.go.tmpl"

	repositoryTmpl, err := fs.ReadFile(repoFilePath)
	if err != nil {
		return nil, err
	}

	methodTmpl, err := fs.ReadFile(methodFilePath)
	if err != nil {
		return nil, err
	}

	transactorTmpl, err := fs.ReadFile(transactorFilePath)
	if err != nil {
		return nil, err
	}

	err = fs.Remove(repoFilePath)
	if err != nil {
		return nil, err
	}

	err = fs.Remove(methodFilePath)
	if err != nil {
		return nil, err
	}

	err = fs.Remove(transactorFilePath)
	if err != nil {
		return nil, err
	}

	for _, adapter := range data.Adapters {
		adapter.AppModule = data.App.Module

		storageFolderPath := fmt.Sprintf("%s/%s", adapterPath, adapter.Storage.GetFolderName())
		if _, err := fs.Stat(storageFolderPath); os.IsNotExist(err) {
			err = fs.Mkdir(storageFolderPath, os.ModeDir)
			if err != nil {
				return nil, err
			}

			generated, err := generate(transactorFilePath, transactorTmpl, adapter)
			if err != nil {
				return nil, err
			}

			err = fs.WriteFile(storageFolderPath+"/transactor.go.tmpl", generated, 0644)
			if err != nil {
				return nil, err
			}
		}

		repoPath := storageFolderPath + "/" + strings.ToLower(adapter.Name)
		if _, err := fs.Stat(repoPath); os.IsNotExist(err) {
			err = fs.Mkdir(repoPath, os.ModeDir)
			if err != nil {
				return nil, err
			}
		}

		generated, err := generate(repoFilePath, repositoryTmpl, adapter)
		if err != nil {
			return nil, err
		}

		err = fs.WriteFile(repoPath+"/repository.go.tmpl", generated, 0644)
		if err != nil {
			return nil, err
		}

		for _, method := range adapter.Methods {
			methodData := struct {
				PkgName    string
				MethodName string
			}{
				PkgName:    adapter.GetPkgName(),
				MethodName: method,
			}

			methodGen, err := generate(methodFilePath, methodTmpl, methodData)
			if err != nil {
				return nil, err
			}

			err = fs.WriteFile(repoPath+"/"+strcase.ToSnake(method)+".go.tmpl", methodGen, 0644)
			if err != nil {
				return nil, err
			}
		}
	}

	return fs, nil
}

func (g *generateAdapterChain) Name() string {
	return "Generate useCase chain"
}

func (g *generateAdapterChain) Rollback() error {
	return nil
}
