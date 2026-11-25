package config

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// namePattern is the rule for the names of services, adapters and methods.
var namePattern = regexp.MustCompile("^[A-Z][A-Za-z]+$")

const nameRule = "a name is two or more letters and starts with a capital letter"

// Validate checks that goro can generate a project from c that builds. init, update and add all
// run it before they write anything, and its errors name the value at fault in quotes. The rules:
//
//   - module, name and work_dir are set;
//   - every storage is one goro supports;
//   - the names of services, adapters and methods follow the name rule;
//   - names are unique across services and adapters, and so are their packages;
//   - every adapter has a storage that storages lists;
//   - each method gets a file of its own in its package, and that file isn't a dep file there, a
//     file goro owns, or a file Go builds only for tests or for one OS or architecture;
//   - the methods of a transactional adapter don't clash with the Transactor its Repository
//     embeds;
//   - a service depends on other services and adapters, the use case on services, each of them
//     listed once.
func (c *Config) Validate() error {
	if c.App.Module == "" || c.App.Name == "" || c.App.WorkDir == "" {
		return fmt.Errorf("module, name and work_dir can't be empty")
	}

	for _, check := range []func() error{
		c.validateStorages,
		c.validateNames,
		c.validateAdapters,
		c.validateServices,
		c.validateUseCase,
	} {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

// SupportedStorages lists the storages goro can generate adapters for, sorted.
func SupportedStorages() []string {
	res := make([]string, 0, len(storagePackages))
	for s := range storagePackages {
		res = append(res, string(s))
	}
	sort.Strings(res)

	return res
}

func isSupported(s Storage) bool {
	_, ok := storagePackages[s]
	return ok
}

func (c *Config) validateStorages() error {
	for _, s := range c.Storages {
		if !isSupported(s) {
			return fmt.Errorf("storage '%s' isn't one goro supports: %s", s, strings.Join(SupportedStorages(), ", "))
		}
	}

	return nil
}

// validateNames checks the names of services and adapters. container.go imports every service
// and adapter package by its package name, the name in lower case, so packages must differ too.
func (c *Config) validateNames() error {
	type owner struct{ kind, name string }
	byName := make(map[string]owner)
	byPkg := make(map[string]owner)

	check := func(o owner, pkg string) error {
		if !namePattern.MatchString(o.name) {
			return fmt.Errorf("%s name '%s' isn't valid: %s", o.kind, o.name, nameRule)
		}
		if other, ok := byName[o.name]; ok {
			if other.kind == o.kind {
				return fmt.Errorf("two %ss are named '%s'", o.kind, o.name)
			}
			return fmt.Errorf("%s '%s' and %s '%s' have the same name; names are unique across services and adapters",
				other.kind, other.name, o.kind, o.name)
		}
		if other, ok := byPkg[pkg]; ok {
			return fmt.Errorf("%s '%s' and %s '%s' get the same package name, %s; names must differ in more than case",
				other.kind, other.name, o.kind, o.name, pkg)
		}
		byName[o.name] = o
		byPkg[pkg] = o

		return nil
	}

	for _, svc := range c.Services {
		if err := check(owner{"service", svc.Name}, svc.GetPkgName()); err != nil {
			return err
		}
	}
	for _, a := range c.Adapters {
		if err := check(owner{"adapter", a.Name}, a.GetPkgName()); err != nil {
			return err
		}
	}

	return nil
}

// transactorMembers are the exported names that the Transactor a transactional adapter's
// Repository embeds brings in: the embedded field and the methods it promotes.
var transactorMembers = []string{"Transactor", "NewTxContext", "InTransaction", "GetConn", "RunTransaction"}

func (c *Config) validateAdapters() error {
	listed := c.Storages.GetMap()

	for _, a := range c.Adapters {
		what := fmt.Sprintf("adapter '%s'", a.Name)
		_, isListed := listed[a.Storage]
		switch {
		case a.Storage == "":
			return fmt.Errorf("%s has no storage; give it one of %s", what, strings.Join(SupportedStorages(), ", "))
		case !isSupported(a.Storage):
			return fmt.Errorf("%s uses storage '%s', which goro doesn't support; use one of %s",
				what, a.Storage, strings.Join(SupportedStorages(), ", "))
		case !isListed:
			return fmt.Errorf("%s uses storage '%s', which storages doesn't list", what, a.Storage)
		}

		pkg := methodPackage{
			what:  what,
			dir:   "internal/adapter/" + a.Storage.GetFolderName() + "/" + a.GetPkgName(),
			owned: "repository.go",
		}
		if a.IsTransactional() {
			pkg.reserved = make(map[string]string, len(transactorMembers))
			for _, m := range transactorMembers {
				pkg.reserved[m] = "clashes with the " + a.Storage.GetFolderName() + ".Transactor that its Repository embeds"
			}
		}
		if err := pkg.check(a.Methods); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) validateServices() error {
	adapters := c.Adapters.GetMap()
	services := c.Services.GetMap()

	for _, svc := range c.Services {
		what := fmt.Sprintf("service '%s'", svc.Name)
		seen := make(map[string]bool, len(svc.Deps))
		for _, d := range svc.Deps {
			_, isAdapter := adapters[d]
			_, isService := services[d]
			switch {
			case d == svc.Name:
				return fmt.Errorf("%s depends on itself", what)
			case !isAdapter && !isService:
				return fmt.Errorf("%s depends on '%s', which is neither a service nor an adapter", what, d)
			case seen[d]:
				return fmt.Errorf("%s lists dependency '%s' twice", what, d)
			}
			seen[d] = true
		}

		pkg := methodPackage{
			what:  what,
			dir:   "internal/service/" + svc.GetPkgName(),
			owned: "service.go",
			deps:  svc.Deps,
		}
		if err := pkg.check(svc.Methods); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) validateUseCase() error {
	services := c.Services.GetMap()

	seen := make(map[string]bool, len(c.UseCase.Deps))
	for _, d := range c.UseCase.Deps {
		if _, ok := services[d]; !ok {
			return fmt.Errorf("the use case depends on '%s', which isn't a service", d)
		}
		if seen[d] {
			return fmt.Errorf("the use case lists dependency '%s' twice", d)
		}
		seen[d] = true
	}

	pkg := methodPackage{
		what:  "the use case",
		dir:   "internal/usecase",
		owned: "usecase.go",
		deps:  c.UseCase.Deps,
	}

	return pkg.check(c.UseCase.Methods)
}

// methodPackage is the package that the methods of the use case, a service or an adapter go
// into, one file each.
type methodPackage struct {
	// what names the owner of the methods in errors, such as service 'Users'.
	what string
	// dir is the package directory, relative to the work dir.
	dir string
	// owned is the file goro owns and regenerates in dir.
	owned string
	// deps are the dependencies whose dep files sit in dir.
	deps []string
	// reserved maps the method names that the generated struct has already to the reason.
	reserved map[string]string
}

func (p methodPackage) check(methods []string) error {
	depFiles := make(map[string]string, len(p.deps))
	for _, d := range p.deps {
		depFiles[DepFile(d)] = d
	}

	files := make(map[string]string, len(methods))
	for _, m := range methods {
		if !namePattern.MatchString(m) {
			return fmt.Errorf("method name '%s' of %s isn't valid: %s", m, p.what, nameRule)
		}

		file := MethodFile(m)
		path := p.dir + "/" + file
		if other, ok := files[file]; ok {
			if other == m {
				return fmt.Errorf("%s lists method '%s' twice", p.what, m)
			}
			return fmt.Errorf("methods '%s' and '%s' of %s both get the file %s", other, m, p.what, path)
		}
		files[file] = m

		if file == p.owned {
			return fmt.Errorf("method '%s' of %s gets the file %s, which goro owns", m, p.what, path)
		}
		if dep, ok := depFiles[file]; ok {
			return fmt.Errorf("method '%s' of %s gets the file %s, which holds the interface of dependency '%s'",
				m, p.what, path, dep)
		}
		if when := buildConstraint(file); when != "" {
			return fmt.Errorf("method '%s' of %s gets the file %s, which Go builds only %s", m, p.what, path, when)
		}
		if why, ok := p.reserved[m]; ok {
			return fmt.Errorf("method '%s' of %s %s", m, p.what, why)
		}
	}

	return nil
}
