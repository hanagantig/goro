package commands_test

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/hanagantig/goro/internal/config"
)

// ownershipConfig covers every kind of generated file: transactional and plain storages, a
// service that depends on adapters and one that depends on a service, and an adapter no service
// uses.
const ownershipConfig = `
app:
  name: demo
  module: demo

storages:
  - mysql
  - pgsqlx
  - http

use_case:
  deps:
    - Users
    - Billing
  methods:
    - SignIn

services:
  - name: Users
    methods:
      - Create
    deps:
      - UserRepo
      - AuditAPI

  - name: Billing
    methods:
      - Charge
    deps:
      - Users

adapters:
  - name: UserRepo
    storage: mysql
    methods:
      - GetByID

  - name: LedgerRepo
    storage: pgsqlx
    methods:
      - Save

  - name: AuditAPI
    storage: http
    methods:
      - Send
`

// TestMarkedFilesAreTheGoroOwnedList pins the files that carry the marker to the list in ADR 0001.
// update overwrites every marked file, so a template that marked a user-owned file would make
// update throw away user code without warning.
func TestMarkedFilesAreTheGoroOwnedList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "goro.yaml")
	if err := os.WriteFile(path, []byte(ownershipConfig), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.NewConfig(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	dir := t.TempDir()
	cfg.App.WorkDir = dir
	initProject(t, cfg)

	marked := make([]string, 0)
	for path, e := range snapshot(t, dir) {
		if !e.isDir && hasMarker(e.content) {
			marked = append(marked, path)
		}
	}
	sort.Strings(marked)

	want := []string{
		"internal/adapter/httprepo/auditapi/repository.go",
		"internal/adapter/mysqlrepo/transactor.go",
		"internal/adapter/mysqlrepo/userrepo/repository.go",
		"internal/adapter/pgsqlxrepo/ledgerrepo/repository.go",
		"internal/adapter/pgsqlxrepo/transactor.go",
		"internal/app/app.go",
		"internal/app/container.go",
		"internal/app/database.go",
		"internal/app/http.go",
		"internal/app/logger.go",
		"internal/service/billing/service.go",
		"internal/service/transactor.go",
		"internal/service/users/service.go",
		"internal/usecase/usecase.go",
	}
	if !reflect.DeepEqual(marked, want) {
		t.Errorf("files with the marker:\n got %q\nwant %q", marked, want)
	}
}
