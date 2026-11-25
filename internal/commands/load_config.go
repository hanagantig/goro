package commands

import "github.com/hanagantig/goro/internal/config"

// LoadConfig reads the goro config at path and asks for the app name and work dir when the
// config leaves them empty.
func LoadConfig(path string) (config.Config, error) {
	cfg, err := config.NewConfig(path)
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
