package config

import (
	"strings"
)

type Service struct {
	AppModule string
	Name      string   `yaml:"name"`
	Methods   []string `yaml:"methods"`
	Deps      []string `yaml:"deps"`
}

type Services []Service

func (s Service) GetConstructorName() string {
	return "NewService"
}

func (s Service) GetPkgName() string {
	return strings.ToLower(s.Name)
}

func (s Service) CheckTransactionalDeps(txAdapterMap map[string]bool) bool {
	for _, depsName := range s.Deps {
		tx, ok := txAdapterMap[depsName]
		if ok && tx {
			return true
		}
	}

	return false
}

func (s Service) GetTransactionalDeps(txAdapterMap map[string]bool) []string {
	res := make([]string, 0, len(s.Deps))
	for _, depsName := range s.Deps {
		tx, ok := txAdapterMap[depsName]
		if ok && tx {
			res = append(res, depsName)
		}
	}

	return res
}

func (s Service) GetNonTransactionalDeps(txAdapterMap map[string]bool) []string {
	res := make([]string, 0, len(s.Deps))
	for _, depsName := range s.Deps {
		tx := txAdapterMap[depsName]
		if !tx {
			res = append(res, depsName)
		}
	}

	return res
}

func (s Services) GetMap() map[string]struct{} {
	res := make(map[string]struct{}, 0)
	for _, svc := range s {
		res[svc.Name] = struct{}{}
	}

	return res
}
