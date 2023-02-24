package config

import "strings"

type Adapter struct {
	Name    string   `yaml:"name"`
	Storage Storage  `yaml:"storage"`
	Methods []string `yaml:"methods"`
}

func (a Adapter) GetPkgName() string {
	return strings.ToLower(a.Name)
}

func (a Adapter) GetConstructorName() string {
	return "NewRepository"
}
