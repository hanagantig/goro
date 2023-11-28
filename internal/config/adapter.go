package config

import (
	"strings"
)

type Adapter struct {
	Name      string   `yaml:"name"`
	Storage   Storage  `yaml:"storage"`
	Methods   []string `yaml:"methods"`
	AppModule string
}

type Adapters []Adapter

func (a Adapter) GetPkgName() string {
	return strings.ToLower(a.Name)
}

func (a Adapter) GetConstructorName() string {
	return "NewRepository"
}

func (a Adapters) GetMap() map[string]struct{} {
	res := make(map[string]struct{}, 0)
	for _, ad := range a {
		res[ad.Name] = struct{}{}
	}

	return res
}

func (a Adapter) IsTransactional() bool {
	return transactionalStorages[a.Storage]
}
