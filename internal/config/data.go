package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v2"
)

type DependencyName string

type Config struct {
	App      App      `yaml:"app"`
	UseCase  UseCase  `yaml:"use_case"`
	Storages Storages `yaml:"storages"`
	Services Services `yaml:"services"`
	Adapters Adapters `yaml:"adapters"`
	Chunks   []Chunk
}

type Chunk struct {
	Name              string
	Scope             string
	ArgName           string
	ReturnType        string
	DefinitionImports string
	BuildImports      string
	InitFunc          string
	Build             string
	Configs           string
}

type App struct {
	Name    string `yaml:"name"`
	Module  string `yaml:"module"`
	WorkDir string `yaml:"work_dir"`
}

type Dependency struct {
	Pkg       string   `yaml:"pkg"`
	Type      string   `yaml:"type"`
	BuildFunc string   `yaml:"build_func"`
	Methods   []string `yaml:"methods"`
	Deps      []string `yaml:"deps"`
}

func (d Dependency) GetPackageName() string {
	path := strings.Split(d.Pkg, "/")
	if len(path) > 0 {
		return path[len(path)-1]
	}

	return ""
}

func NewConfig(pathToFile string) (Config, error) {
	if pathToFile == "" {
		return Config{}, nil
	}

	return LoadDataFromYaml(pathToFile)
}

func LoadDataFromYaml(pathToFile string) (Config, error) {
	var data Config

	_, err := os.Stat(pathToFile)
	if err != nil {
		return data, err
	}

	fileContents, _ := ioutil.ReadFile(pathToFile)

	err = yaml.Unmarshal(fileContents, &data)

	return data, err
}

func (c *Config) AskAndSetName() error {
	if c.App.Name != "" {
		return nil
	}

	name, err := c.askName()
	if err != nil {
		return err
	}

	c.App.Name = name
	return nil
}

func (c *Config) askName() (string, error) {
	return (&promptui.Prompt{
		Label: "Enter an app name",
		Validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("config name can't be empty")
			}
			return nil
		},
	}).Run()
}

func (c *Config) AskAndSetWorkDir() error {
	if c.App.WorkDir != "" {
		return nil
	}

	wd, err := c.askWorkDir()
	if err != nil {
		return err
	}

	if wd == "" {
		wd, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	c.App.WorkDir = wd
	return nil
}

func (c *Config) askWorkDir() (string, error) {
	return (&promptui.Prompt{
		Label: "Enter an app work dir or leave it empty to use current directory",
		// TODO: validate path
	}).Run()
}
