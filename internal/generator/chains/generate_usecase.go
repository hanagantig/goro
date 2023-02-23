package chains

import (
	entity "github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/pkg/afero"
	"github.com/iancoleman/strcase"
)

type generateUseCaseChain struct{}

func NewGenerateUseCaseChain() *generateUseCaseChain {
	return &generateUseCaseChain{}
}

func (g *generateUseCaseChain) Apply(fs *afero.Afero, data entity.Config) (*afero.Afero, error) {
	useCasePath := "/internal/usecase"
	useCaseFilePath := "/internal/usecase/usecase.go.tmpl"
	methodFilePath := "/internal/usecase/method.go.tmpl"

	useCaseTmpl, err := fs.ReadFile(useCaseFilePath)
	if err != nil {
		return nil, err
	}

	methodTmpl, err := fs.ReadFile(methodFilePath)
	if err != nil {
		return nil, err
	}

	err = fs.Remove(useCaseFilePath)
	if err != nil {
		return nil, err
	}

	err = fs.Remove(methodFilePath)
	if err != nil {
		return nil, err
	}

	generated, err := generate(useCaseFilePath, useCaseTmpl, data.UseCase)
	if err != nil {
		return nil, err
	}

	err = fs.WriteFile(useCasePath+"/usecase.go.tmpl", generated, 0644)
	if err != nil {
		return nil, err
	}

	for _, method := range data.UseCase.Methods {
		methodGen, err := generate(methodFilePath, methodTmpl, method)
		if err != nil {
			return nil, err
		}

		err = fs.WriteFile(useCasePath+"/"+strcase.ToSnake(method)+".go.tmpl", methodGen, 0644)
		if err != nil {
			return nil, err
		}
	}

	return fs, nil
}

func (g *generateUseCaseChain) Name() string {
	return "Generate useCase chain"
}

func (g *generateUseCaseChain) Rollback() error {
	return nil
}
