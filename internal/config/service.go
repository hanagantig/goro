package config

import "strings"

type Service struct {
	AppModule string
	Name      string   `yaml:"name"`
	Methods   []string `yaml:"methods"`
	Deps      []string `yaml:"deps"`
}

func (s Service) GetConstructorName() string {
	return "NewService"
}

func (s Service) GetPkgName() string {
	return strings.ToLower(s.Name)
}
