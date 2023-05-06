package config

import (
	"fmt"
	"regexp"
)

var methodName = regexp.MustCompile("^[A-Z][A-Za-z]+$")

func (c *Config) Validate() error {
	if c.App.Module == "" || c.App.Name == "" || c.App.WorkDir == "" {
		return fmt.Errorf("module, name and work_dir can't be empty")
	}

	storeMap := c.Storages.GetMap()
	adapterMap := c.Adapters.GetMap()
	serviceMap := c.Services.GetMap()

	for _, adapter := range c.Adapters {
		if _, ok := storeMap[adapter.Storage]; !ok {
			return fmt.Errorf("not supported adapter storage type '%v'", adapter.Storage)
		}

		for _, m := range adapter.Methods {
			if !methodName.Match([]byte(m)) {
				return fmt.Errorf("adapter method name '%v' should start with capitalized letter", m)
			}
		}
	}

	for _, svc := range c.Services {
		for _, d := range svc.Deps {
			_, adapterOk := adapterMap[d]
			_, serviceOk := serviceMap[d]

			if !(adapterOk || serviceOk) {
				return fmt.Errorf("service dependancy '%v' should be adapter or other service", d)
			}
		}

		for _, m := range svc.Methods {
			if !methodName.Match([]byte(m)) {
				return fmt.Errorf("service method name '%v' should start with capitalized letter", m)
			}
		}
	}

	for _, ucDep := range c.UseCase.Deps {
		if _, ok := serviceMap[ucDep]; !ok {
			return fmt.Errorf("usecase dependancy '%v' should be a services", ucDep)
		}
	}

	for _, ucm := range c.UseCase.Methods {
		if !methodName.Match([]byte(ucm)) {
			return fmt.Errorf("usecase method name '%v' should start with capitalized letter", ucm)
		}
	}

	return nil
}
