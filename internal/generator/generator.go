package generator

import (
	"fmt"
	"github.com/spf13/afero"
	entity "goro/internal/config"
	"io/fs"
	"os"
	"sync"
)

type Chain interface {
	Name() string
	Apply(fs *afero.Afero) (*afero.Afero, error)
	Rollback() error
}

type Generator struct {
	mu       sync.RWMutex
	config   entity.Config
	chains   []Chain
	skeleton skeleton
}

func NewGenerator() *Generator {
	return &Generator{
		chains:   make([]Chain, 0),
		skeleton: singleServiceSkeleton,
	}
}

func (g *Generator) AddChain(ch Chain) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.chains = append(g.chains, ch)
}

func (g *Generator) getTemplateFS() (*afero.Afero, error) {
	mfs := afero.NewMemMapFs()
	afs := &afero.Afero{Fs: mfs}

	err := fs.WalkDir(g.skeleton.template, g.skeleton.root,
		func(path string, d fs.DirEntry, err error) error {
			savePath := path[len(g.skeleton.root):]
			if savePath == "" {
				return nil
			}

			if d.IsDir() {
				err = afs.MkdirAll(savePath, os.FileMode(0775))
				if err != nil {
					return err
				}
				return nil
			}

			content, err := g.skeleton.template.ReadFile(savePath)
			if err != nil {
				return err
			}

			err = afs.WriteFile(savePath, content, os.FileMode(0775))
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
