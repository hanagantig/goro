package config

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

var storagePackages = map[string]string{
	"mysql":  "\"database/sql\"",
	"mysqlx": "\"github.com/jmoiron/sqlx\"",
}

type DependencyName string
type StorageName string
type StorageList map[StorageName]Storage

func (s StorageName) String() string {
	return string(s)
}

func (s DependencyName) String() string {
	return string(s)
}

type Config struct {
	App          App                           `yaml:"app"`
	UseCase      UseCase                       `yaml:"use_case"`
	Storages     StorageList                   `yaml:"storages"`
	Dependencies map[DependencyName]Dependency `yaml:"dependencies"`
}

type UseCase struct {
	Pkg       string   `yaml:"pkg"`
	Type      string   `yaml:"type"`
	BuildFunc string   `yaml:"build_func"`
	Deps      []string `yaml:"deps"`
}

type App struct {
	Name    string `yaml:"name"`
	Module  string `yaml:"module"`
	WorkDir string `yaml:"work_dir"`
}

type Storage struct {
	Type       string `yaml:"type"`
	Connection string `yaml:"connection"`
}

type Dependency struct {
	Pkg       string   `yaml:"pkg"`
	Type      string   `yaml:"type"`
	BuildFunc string   `yaml:"build_func"`
	Deps      []string `yaml:"deps"`
}

func (s *Storage) GetPackage() string {
	return storagePackages[s.Type]
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

func (d *Config) Validate() error {
	if d.App.Module == "" || d.App.Name == "" || d.App.WorkDir == "" {
		return fmt.Errorf("module, name and work_dir can't be empty")
	}
	for depName, dep := range d.Dependencies {
		if _, ok := d.Storages[StorageName(depName)]; ok {
			return fmt.Errorf("dependency \"%v\" has the same name with storage", depName)
		}

		for _, dpName := range dep.Deps {
			if depName == DependencyName(dpName) {
				return fmt.Errorf("dependency \"%v\" can't depend on self", depName)
			}

			_, okDep := d.Dependencies[DependencyName(dpName)]
			_, okStore := d.Storages[StorageName(dpName)]
			if !okDep && !okStore {
				return fmt.Errorf("undefined dep name \"%v\" in \"%v\" dependency", dpName, depName)
			}
		}
	}

	return nil
}

func (d *Config) AskAndSetName() error {
	if d.App.Name != "" {
		return nil
	}

	name, err := d.askName()
	if err != nil {
		return err
	}

	d.App.Name = name
	return nil
}

func (d *Config) askName() (string, error) {
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

func (d *Config) AskAndSetWorkDir() error {
	if d.App.WorkDir != "" {
		return nil
	}

	wd, err := d.askWorkDir()
	if err != nil {
		return err
	}

	if wd == "" {
		wd, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	d.App.WorkDir = wd
	return nil
}

func (d *Config) askWorkDir() (string, error) {
	return (&promptui.Prompt{
		Label: "Enter an app work dir or leave it empty to use current directory",
		// TODO: validate path
	}).Run()
}
