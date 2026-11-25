package commands

import (
	"errors"

	"github.com/hanagantig/goro/internal/config"
)

// The goro add functions change the goro.yaml at configPath, the project's source of truth
// (ADR 0002), and then run Update with opts on the result. They share these rules:
//
//   - they validate the change against the whole config first; a failure returns an error that
//     names the value at fault and leaves both goro.yaml and the code as they are;
//   - they edit goro.yaml in place and keep its comments and key order;
//   - a relative work_dir in goro.yaml counts from the directory that holds it, as in LoadConfig;
//   - under opts.DryRun they print what Update would do and write nothing, goro.yaml included.
//
// They return the Report of that Update.

var errAddNotImplemented = errors.New("not implemented")

// AddUseCase adds method to use_case.methods, which creates its file in internal/usecase.
// Not implemented yet: returns an error.
func AddUseCase(configPath, method string, opts Options) (Report, error) {
	return Report{}, errAddNotImplemented
}

// AddService adds svc to services with its methods and deps: Update creates its package with a
// file per method and a dep file per dependency, and wires it in the container. Each dependency
// must be a service or an adapter that goro.yaml lists.
// Not implemented yet: returns an error.
func AddService(configPath string, svc config.Service, opts Options) (Report, error) {
	return Report{}, errAddNotImplemented
}

// AddAdapter adds adapter to adapters with its storage and methods, and adds the storage to
// storages when goro.yaml doesn't list it yet. The storage must be one goro supports.
// Not implemented yet: returns an error.
func AddAdapter(configPath string, adapter config.Adapter, opts Options) (Report, error) {
	return Report{}, errAddNotImplemented
}

// AddMethod adds a method to a service or an adapter. ref is "<Name>.<Method>", such as
// "Billing.Void". It fails when Name is both a service and an adapter, or neither.
// Not implemented yet: returns an error.
func AddMethod(configPath, ref string, opts Options) (Report, error) {
	return Report{}, errAddNotImplemented
}
