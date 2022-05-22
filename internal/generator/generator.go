package generator

import (
	"fmt"
	"sync"
)

type Chain interface {
	Name() string
	Apply() error
	Rollback() error
}

type Generator struct {
	mu     sync.RWMutex
	chains []Chain
}

func NewGenerator() *Generator {
	return &Generator{
		chains: make([]Chain, 0),
	}
}

func (g *Generator) AddChain(ch Chain) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.chains = append(g.chains, ch)
}

func (g *Generator) Generate() error {
	for k, ch := range g.chains {
		fmt.Printf("chain #%d: %s \n", k+1, ch.Name())
		err := ch.Apply()
		if err != nil {
			_ = ch.Rollback()
			return err
		}
	}

	return nil
}
