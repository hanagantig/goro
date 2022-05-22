package chains

import (
	cp "github.com/otiai10/copy"
	entity "goro/internal/entity"
)

type basementChain struct {
}

func NewBasementChain(data entity.AppData) *basementChain {
	return &basementChain{}
}

func (b *basementChain) Apply() error {
	opt := cp.Options{
		Skip: func(src string) (bool, error) {
			return src == ".DS_Store", nil
		},
	}
	return cp.Copy("/Users/hanagantig/projects/goro/internal/generator/templates/app/", "/Users/hanagantig/tmp/gorotest", opt)
}

func (b *basementChain) Name() string {
	return "Basement chain"
}

func (b *basementChain) Rollback() error {
	return nil
}
