package commands

import (
	"errors"
	"io"
)

// Options control one run of Init or Update.
type Options struct {
	// SkipGoMod leaves out the `go mod init` and `go mod tidy` steps. They run the go tool and
	// need the network, so in-process tests set it.
	SkipGoMod bool

	// DryRun makes Update print and return its actions without writing anything.
	// Not implemented yet.
	DryRun bool

	// Out receives the printed report. Nil means os.Stdout.
	// Not used yet.
	Out io.Writer
}

// ErrProjectExists is returned by Init when the work dir already holds a goro project.
// Not returned yet.
var ErrProjectExists = errors.New("the work dir already holds a goro project; run goro update instead")

// Action is what a run of Update does, or would do, with one path in the work dir.
type Action string

const (
	// ActionCreate: the file is missing, goro writes it.
	ActionCreate Action = "create"
	// ActionOverwrite: a generated file whose content changes, goro rewrites it.
	ActionOverwrite Action = "overwrite"
	// ActionUnchanged: a generated file that is already up to date, goro doesn't write it.
	ActionUnchanged Action = "unchanged"
	// ActionSkip: a user-owned file that exists, goro never touches it.
	ActionSkip Action = "skip"
	// ActionSkipUnmarked: a file at a generated path without the generated marker. It belongs
	// to the user; goro doesn't write it and warns about it.
	ActionSkipUnmarked Action = "skip-unmarked"
	// ActionOrphan: code that goro.yaml doesn't list (a use case method file, a service or
	// adapter package, a dep_*.go file). goro keeps it and reports it.
	ActionOrphan Action = "orphan"
)

// FileAction is one entry of a Report.
type FileAction struct {
	// Path is relative to the work dir, with forward slashes. For an orphaned package it is
	// the package directory, such as internal/service/mailer.
	Path   string
	Action Action
}

// Report lists what Update did, or would do under DryRun: one entry per file goro generates and
// per orphan. Update doesn't fill it yet.
type Report struct {
	Files []FileAction
}
