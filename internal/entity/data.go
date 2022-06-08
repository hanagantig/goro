package entity

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
)

type DependencyName string

type AppData struct {
	Name         string
	Module       string
	WorkDir      string
	Databases    []Database
	Dependencies map[DependencyName]Dependency
}

type Database struct {
	Type       string
	Name       string
	Connection string
}

type Dependency struct {
	Pkg       string
	Type      string
	BuildFunc string
	Deps      []DependencyName
}

func (d *AppData) AskAndSetName() error {
	name, err := d.askName()
	if err != nil {
		return err
	}

	d.Name = name
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
	d.WorkDir = wd
	return nil
}

func (d *AppData) askWorkDir() (string, error) {
	return (&promptui.Prompt{
		Label: "Enter an app work dir or leave it empty to use current directory",
		// TODO: validate path
	}).Run()
}
