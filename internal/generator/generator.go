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

// Generate renders the templates through every chain into an in-memory file system and returns
// it. Paths in it start at "/", which stands for the work dir. Generate doesn't touch the disk;
// the caller decides what to write.
func (g *Generator) Generate() (*afero.Afero, error) {
	fs, err := g.getTemplateFS()
	if err != nil {
		return nil, err
	}

	err = g.loadChunks()
	if err != nil {
		return nil, err
	}

	for _, ch := range g.chains {
		fs, err = ch.Apply(fs, g.config)
		if err != nil {
			_ = ch.Rollback()
			return nil, fmt.Errorf("%s: %w", ch.Name(), err)
		}
	}

	return fs, nil
}
