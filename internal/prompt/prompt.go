package prompt

import (
	"fmt"
	"github.com/manifoldco/promptui"
)

func AskAppName() (string, error) {
	// TODO: use it in AppData and check if there is such flag
	return (&promptui.Prompt{
		Label: "Enter an app name",
		Validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("app name can't be empty")
			}
			return nil
		},
	}).Run()
}
