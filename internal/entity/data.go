package entity

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

type AppData struct {
	Name      string
	Module    string
	Databases []Database
}

type Database struct {
	Type string
	Name string
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
