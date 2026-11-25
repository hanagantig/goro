package chains

import (
	"github.com/iancoleman/strcase"

	entity "github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/internal/generator"
	"github.com/hanagantig/goro/pkg/afero"
)

// DepFilePrefix starts the name of every dependency interface file, dep_<snake name>.go.
const DepFilePrefix = "dep_"

type generateDepsChain struct{}

// NewGenerateDepsChain renders one user-owned dep_<snake name>.go file per dependency of the use
// case and of each service. Each file declares the interface the generated struct uses for that
// dependency. Run it before the generate code chain, which would otherwise render the template
// itself.
func NewGenerateDepsChain() *generateDepsChain {
	return &generateDepsChain{}
}

type depData struct {
	PkgName       string
	Interface     string
	Dep           string
	Transactional bool
	AppModule     string
}

func (g *generateDepsChain) Apply(fs *afero.Afero, data entity.Config) (*afero.Afero, error) {
	const depFilePath = "/internal/dep.go.tmpl"

	depTmpl, err := fs.ReadFile(depFilePath)
	if err != nil {
		return nil, err
	}

	err = fs.Remove(depFilePath)
	if err != nil {
		return nil, err
	}

	write := func(dir, pkgName, dep string, transactional bool) error {
		generated, err := generate(depFilePath, depTmpl, depData{
			PkgName:       pkgName,
			Interface:     generator.ToPrivateName(dep),
			Dep:           dep,
			Transactional: transactional,
			AppModule:     data.App.Module,
		})
		if err != nil {
			return err
		}

		return fs.WriteFile(dir+"/"+DepFilePrefix+strcase.ToSnake(dep)+".go", generated, 0644)
	}

	for _, dep := range data.UseCase.Deps {
		err = write("/internal/usecase", "usecase", dep, false)
		if err != nil {
			return nil, err
		}
	}

	trxMap := data.Adapters.GetTransactionalMap()
	for _, svc := range data.Services {
		for _, dep := range svc.Deps {
			err = write("/internal/service/"+svc.GetPkgName(), svc.GetPkgName(), dep, trxMap[dep])
			if err != nil {
				return nil, err
			}
		}
	}

	return fs, nil
}

func (g *generateDepsChain) Name() string {
	return "Generate dependency interfaces"
}

func (g *generateDepsChain) Rollback() error {
	return nil
}
