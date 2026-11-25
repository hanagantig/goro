package commands

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/hanagantig/goro/internal/config"
)

// LoadConfig reads the goro config at path, with a relative work_dir counted from the directory
// that holds it, and asks for the app name and work dir when the config leaves them empty.
func LoadConfig(path string) (config.Config, error) {
	cfg, err := config.NewConfig(path)
	if errors.Is(err, fs.ErrNotExist) {
		return config.Config{}, fmt.Errorf("%s doesn't exist; run goro in the project root, where init saves goro.yaml, "+
			"or point --config to a goro.yaml", path)
	}
	if err != nil {
		return config.Config{}, err
	}

	err = cfg.AskAndSetName()
	if err != nil {
		return config.Config{}, err
	}

	err = cfg.AskAndSetWorkDir()
	if err != nil {
		return config.Config{}, err
	}

	return cfg, nil
}
