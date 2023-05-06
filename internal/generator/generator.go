package generator

import (
	"fmt"
	entity "github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/pkg/afero"
	"io/fs"
	"os"
	"sync"
)

type Chain interface {
	Name() string
	Apply(fs *afero.Afero, data entity.Config) (*afero.Afero, error)
	Rollback() error
}

type Generator struct {
	mu       sync.RWMutex
	config   entity.Config
	chains   []Chain
	skeleton skeleton
}

func NewGenerator(config entity.Config) *Generator {
	return &Generator{
		chains:   make([]Chain, 0),
		skeleton: singleServiceSkeleton,
		config:   config,
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

			content, err := g.skeleton.template.ReadFile(path)
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

func (g *Generator) loadChunks() error {
	for _, name := range g.config.Storages {
		ch, ok := supportedChunks[name.String()]
		if !ok {
			return fmt.Errorf("%w: %v", UnsupportedChunkErr, name)
		}

		g.config.Chunks = append(g.config.Chunks, ch)
	}

	return nil
}

func (g *Generator) Generate() error {
	fs, err := g.getTemplateFS()
	if err != nil {
		return err
	}

	err = g.loadChunks()
	if err != nil {
		return err
	}

	for k, ch := range g.chains {
		fmt.Printf("chain #%d: %s \n", k+1, ch.Name())

		fs, err = ch.Apply(fs, g.config)
		if err != nil {
			fmt.Printf("%s fail: %v", ch.Name(), err)
			_ = ch.Rollback()
			return err
		}
	}

	return nil
}
