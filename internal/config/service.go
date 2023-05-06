package config

import "strings"

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

func (s Services) GetMap() map[string]struct{} {
	res := make(map[string]struct{}, 0)
	for _, svc := range s {
		res[svc.Name] = struct{}{}
	}

	return res
}
