package generator

import (
	"embed"
	"fmt"
	"github.com/spf13/afero"
	entity "goro/internal/config"
	"io/fs"
	"os"
	"sync"
)

//go:embed templates/app
var appTmplFs embed.FS

type Chain interface {
	Name() string
	Apply(fs *afero.Afero) (*afero.Afero, error)
	Rollback() error
}

type Generator struct {
	mu       sync.RWMutex
	config   entity.Config
	chains   []Chain
	skeleton embed.FS
}

func NewGenerator() *Generator {
	return &Generator{
		chains:   make([]Chain, 0),
		skeleton: appTmplFs,
	}
}

func (g *Generator) GetAppTemplate() embed.FS {
	return appTmplFs
}

func (g *Generator) AddChain(ch Chain) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.chains = append(g.chains, ch)
}

func (g *Generator) getTemplateFS() (*afero.Afero, error) {
	mfs := afero.NewMemMapFs()
	afs := &afero.Afero{Fs: mfs}

	err := fs.WalkDir(g.skeleton, "templates/app",
		func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				err = afs.MkdirAll(path, os.FileMode(0775))
				if err != nil {
					return err
				}
				return nil
			}

			content, err := g.skeleton.ReadFile(path)
			if err != nil {
				return err
			}

			err = afs.WriteFile(path, content, os.FileMode(0775))
			if err != nil {
				return err
			}

			return nil
		},
	)

	return afs, err
}

func (g *Generator) Generate() error {

	fs, err := g.getTemplateFS()
	if err != nil {
		return err
	}

	for k, ch := range g.chains {
		fmt.Printf("chain #%d: %s \n", k+1, ch.Name())
		_, err = fs.ReadDir("cmd")

		fs, err = ch.Apply(fs)
		if err != nil {
			fmt.Printf("%s fail: %v", ch.Name(), err)
			_ = ch.Rollback()
			return err
		}
	}

	return nil
}
