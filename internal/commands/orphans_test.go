package commands_test

import (
	"reflect"
	"testing"

	"github.com/hanagantig/goro/internal/config"
)

// TestUpdateOrphanRules covers the orphan rules that TestUpdateKeepsUnlistedCodeAndReportsOrphans
// leaves out: a whole storage folder, removed method files in listed service and adapter
// packages, the dep file of a removed service dependency, helper files, and what doesn't count.
func TestUpdateOrphanRules(t *testing.T) {
	dir := t.TempDir()

	// Init with an http storage and adapter on top of baseConfig.
	initCfg := newConfig(t, dir)
	initCfg.Storages = append(initCfg.Storages, "http")
	initCfg.Adapters = append(initCfg.Adapters, config.Adapter{Name: "PaymentsAPI", Storage: "http", Methods: []string{"Charge"}})
	initProject(t, initCfg)

	handMade := map[string]string{
		// A helper file in a listed package is code goro.yaml doesn't describe.
		"internal/usecase/errors.go": "package usecase\n",
		// Test files, non-Go files and subdirectories of internal/usecase don't count.
		"internal/usecase/sign_in_test.go":                      "package usecase\n",
		"internal/usecase/README.md":                            "notes\n",
		"internal/usecase/mocks/users.go":                       "package mocks\n",
		"internal/service/users/create_test.go":                 "package users\n",
		"internal/adapter/mysqlrepo/userrepo/get_by_id_test.go": "package userrepo\n",
	}
	for path, content := range handMade {
		writeFile(t, dir, path, content)
	}
	before := snapshot(t, dir)

	// goro.yaml: back to baseConfig without http, and Users loses its method and its UserRepo
	// dependency, UserRepo its method.
	cfg := newConfig(t, dir)
	for i := range cfg.Services {
		if cfg.Services[i].Name == "Users" {
			cfg.Services[i].Methods = nil
			cfg.Services[i].Deps = nil
		}
	}
	cfg.Adapters[0].Methods = nil

	report := update(t, cfg)

	want := []string{
		"internal/adapter/httprepo",
		"internal/adapter/mysqlrepo/userrepo/get_by_id.go",
		"internal/service/users/create.go",
		"internal/service/users/dep_user_repo.go",
		"internal/usecase/errors.go",
	}
	if got := orphans(report); !reflect.DeepEqual(got, want) {
		t.Errorf("orphans:\n got %q\nwant %q", got, want)
	}
	assertNothingDeleted(t, before, snapshot(t, dir))
}
