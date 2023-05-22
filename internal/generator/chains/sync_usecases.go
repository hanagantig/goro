package chains

import (
	entity "github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/pkg/afero"
	"github.com/iancoleman/strcase"
	"os"
	"path/filepath"
	"strings"
)

type syncUseCaseChain struct {
	data entity.Config
}

func NewSyncUseCaseChain() *syncUseCaseChain {
	return &syncUseCaseChain{}
}

func (g *syncUseCaseChain) Apply(fs *afero.Afero, data entity.Config) (*afero.Afero, error) {

	useCasePath := filepath.Join(data.App.WorkDir, "/internal/usecase")

	var loadedUseCaseFiles []string
	for _, ucm := range data.UseCase.Methods {
		loadedUseCaseFiles = append(loadedUseCaseFiles, strcase.ToSnake(ucm)+".go")
	}

	useCaseFiles, err := os.ReadDir(useCasePath)
	for _, ucf := range useCaseFiles {
		if !contains(loadedUseCaseFiles, strings.ToLower(ucf.Name())) {
			err = os.RemoveAll(filepath.Join(useCasePath, ucf.Name()))
			if err != nil {
				return fs, err
			}
		}
	}

	return fs, nil
}

func (g *syncUseCaseChain) Name() string {
	return "Synchronize usecases"
}

func (g *syncUseCaseChain) Rollback() error {
	return nil
}
