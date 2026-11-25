package commands

import (
	"fmt"
	"strings"

	"github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/internal/yamledit"
)

// The goro add functions change the goro.yaml at configPath, the project's source of truth
// (ADR 0002), and then run Update with opts on the result. They share these rules:
//
//   - they validate the change against the whole config first; a failure returns an error that
//     names the value at fault, quoted or qualified ('Billing', Users.Void), and leaves both
//     goro.yaml and the code as they are;
//   - they edit goro.yaml in place by inserting lines only, so every existing line stays as it
//     is; a missing key is inserted with its items, and a flow list ([a, b]) is refused with an
//     error that names the file and the list by its path, such as use_case.methods or
//     adapters.UserRepo.methods;
//   - a relative work_dir in goro.yaml counts from the directory that holds it, as in LoadConfig;
//   - they never prompt: a goro.yaml without app.name or app.work_dir is an error that says what
//     to set;
//   - under opts.DryRun they print what Update would do and write nothing, goro.yaml included.
//
// They return the Report of that Update. goro.yaml is written first and isn't in the report.

// AddUseCase adds method to use_case.methods, which creates its file in internal/usecase.
func AddUseCase(configPath, method string, opts Options) (Report, error) {
	g, err := readGoroYaml(configPath)
	if err != nil {
		return Report{}, err
	}

	g.cfg.UseCase.Methods = append(g.cfg.UseCase.Methods, method)

	return g.save(opts, func() error {
		return g.appendTo(g.root, "", []string{"use_case", "methods"}, method)
	})
}

// AddService adds svc to services with its methods and deps: Update creates its package with a
// file per method and a dep file per dependency, and wires it in the container. Each dependency
// must be a service or an adapter that goro.yaml lists.
func AddService(configPath string, svc config.Service, opts Options) (Report, error) {
	g, err := readGoroYaml(configPath)
	if err != nil {
		return Report{}, err
	}

	svc = config.Service{Name: svc.Name, Methods: svc.Methods, Deps: svc.Deps}
	g.cfg.Services = append(g.cfg.Services, svc)

	return g.save(opts, func() error {
		item := yamledit.Map{
			{Key: "name", Value: svc.Name},
			{Key: "methods", Value: list(svc.Methods)},
			{Key: "deps", Value: list(svc.Deps)},
		}
		return g.appendTo(g.root, "", []string{"services"}, item)
	})
}

// AddAdapter adds adapter to adapters with its storage and methods, and adds the storage to
// storages when goro.yaml doesn't list it yet. The storage must be one goro supports.
func AddAdapter(configPath string, adapter config.Adapter, opts Options) (Report, error) {
	g, err := readGoroYaml(configPath)
	if err != nil {
		return Report{}, err
	}

	a := config.Adapter{Name: adapter.Name, Storage: adapter.Storage, Methods: adapter.Methods}
	_, listed := g.cfg.Storages.GetMap()[a.Storage]
	newStorage := a.Storage != "" && !listed
	if newStorage {
		g.cfg.Storages = append(g.cfg.Storages, a.Storage)
	}
	g.cfg.Adapters = append(g.cfg.Adapters, a)

	return g.save(opts, func() error {
		if newStorage {
			if err := g.appendTo(g.root, "", []string{"storages"}, a.Storage.String()); err != nil {
				return err
			}
		}
		item := yamledit.Map{
			{Key: "name", Value: a.Name},
			{Key: "storage", Value: a.Storage.String()},
			{Key: "methods", Value: list(a.Methods)},
		}
		return g.appendTo(g.root, "", []string{"adapters"}, item)
	})
}

// AddMethod adds a method to a service or an adapter. ref is "<Name>.<Method>", such as
// "Billing.Void"; any other form is an error. It fails when Name is both a service and an
// adapter, or neither.
func AddMethod(configPath, ref string, opts Options) (Report, error) {
	name, method, ok := strings.Cut(ref, ".")
	if !ok || name == "" || method == "" {
		return Report{}, fmt.Errorf("'%s' isn't <Name>.<Method>, such as Billing.Void", ref)
	}

	g, err := readGoroYaml(configPath)
	if err != nil {
		return Report{}, err
	}

	service, adapter := -1, -1
	for i, s := range g.cfg.Services {
		if s.Name == name {
			service = i
		}
	}
	for i, a := range g.cfg.Adapters {
		if a.Name == name {
			adapter = i
		}
	}

	var section string
	switch {
	case service >= 0 && adapter >= 0:
		return Report{}, fmt.Errorf("'%s' is both a service and an adapter in %s; rename one of them", name, g.path)
	case service >= 0:
		section = "services"
		g.cfg.Services[service].Methods = append(g.cfg.Services[service].Methods, method)
	case adapter >= 0:
		section = "adapters"
		g.cfg.Adapters[adapter].Methods = append(g.cfg.Adapters[adapter].Methods, method)
	default:
		return Report{}, fmt.Errorf("'%s' is neither a service nor an adapter in %s", name, g.path)
	}

	return g.save(opts, func() error {
		item, label, err := g.item(section, name)
		if err != nil {
			return err
		}
		return g.appendTo(item, label, []string{"methods"}, method)
	})
}

// list returns names as a list to insert.
func list(names []string) yamledit.List {
	res := make(yamledit.List, 0, len(names))
	for _, n := range names {
		res = append(res, n)
	}

	return res
}
