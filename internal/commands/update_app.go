package commands

import (
	"github.com/hanagantig/goro/internal/config"
)

// Update brings the project in cfg.App.WorkDir in line with cfg (ADR 0001). It regenerates the
// files that carry goro's marker, creates missing files, leaves every other file alone, deletes
// nothing, and reports what it did with each file and the orphans.
func Update(cfg config.Config, opts Options) (Report, error) {
	err := cfg.Validate()
	if err != nil {
		return Report{}, err
	}

	return run(cfg, opts, "update", false, nil)
}
