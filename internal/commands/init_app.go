package commands

import (
	"fmt"

	"github.com/hanagantig/goro/internal/config"
)

// Init generates a new project from cfg into cfg.App.WorkDir. It writes through the same planner
// as Update, so it never overwrites a file that is already there, and it refuses a work dir that
// already holds a goro project with ErrProjectExists. A work dir that holds other files, such as
// goro.yaml, is fine.
func Init(cfg config.Config, opts Options) error {
	err := cfg.Validate()
	if err != nil {
		return err
	}

	exists, err := isGoroProject(cfg.App.WorkDir)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("init %s: %w", cfg.App.WorkDir, ErrProjectExists)
	}

	_, err = run(cfg, opts, "init", true)

	return err
}
