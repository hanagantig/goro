package entity

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type DependencyName string
type StorageName string

func (s StorageName) String() string {
	return string(s)
}

func (s DependencyName) String() string {
	return string(s)
}

type AppData struct {
	App          App                           `yaml:"app"`
	Storages     map[StorageName]Storage       `yaml:"storages"`
	Dependencies map[DependencyName]Dependency `yaml:"dependencies"`
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

func LoadDataFromYaml(pathToFile string) (AppData, error) {
	var data AppData

	_, err := os.Stat(pathToFile)
	if err != nil {
		return data, err
	}

	fileContents, _ := ioutil.ReadFile(pathToFile)

	err = yaml.Unmarshal(fileContents, &data)

	return data, err
}

func (d *AppData) Validate() error {
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

func (d *AppData) AskAndSetName() error {
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

func (d *AppData) askName() (string, error) {
	return (&promptui.Prompt{
		Label: "Enter an app name",
		Validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("entity name can't be empty")
			}
			return nil
		},
	}).Run()
}

func (d *AppData) AskAndSetWorkDir() error {
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

func (d *AppData) askWorkDir() (string, error) {
	return (&promptui.Prompt{
		Label: "Enter an app work dir or leave it empty to use current directory",
		// TODO: validate path
	}).Run()
}
