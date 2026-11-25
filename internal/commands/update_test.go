package commands_test

import (
	"bytes"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/hanagantig/goro/internal/commands"
	"github.com/hanagantig/goro/internal/config"
)

// Rules under test: ADR 0001, generated and user-owned files; update never deletes.

func TestUpdateKeepsHandEditsInMethodFiles(t *testing.T) {
	dir := t.TempDir()
	initProject(t, newConfig(t, dir))

	methodFiles := []string{
		"internal/usecase/sign_in.go",
		"internal/service/users/create.go",
		"internal/adapter/mysqlrepo/userrepo/get_by_id.go",
	}
	edited := make(map[string]string, len(methodFiles))
	for _, path := range methodFiles {
		edited[path] = handEdit(readFile(t, dir, path))
		writeFile(t, dir, path, edited[path])
	}

	update(t, newConfig(t, dir))

	for _, path := range methodFiles {
		t.Run(path, func(t *testing.T) {
			if !exists(dir, path) {
				t.Fatalf("%s was deleted", path)
			}
			if readFile(t, dir, path) != edited[path] {
				t.Errorf("update overwrote the hand edits in %s", path)
			}
		})
	}
}

func TestUpdateKeepsUnlistedCodeAndReportsOrphans(t *testing.T) {
	dir := t.TempDir()
	initProject(t, newConfig(t, dir))

	handMade := map[string]string{
		"internal/service/mailer/mailer.go":             "package mailer\n\nfunc Send() error { return nil }\n",
		"internal/adapter/mysqlrepo/legacyrepo/find.go": "package legacyrepo\n\nfunc Find() error { return nil }\n",
		"internal/usecase/dep_billing.go":               "package usecase\n\ntype billing interface {\n\tCharge() error\n}\n",
	}
	for path, content := range handMade {
		writeFile(t, dir, path, content)
	}
	// Code goro made for what the edited goro.yaml no longer lists.
	keep := map[string]string{
		"internal/usecase/sign_up.go":        readFile(t, dir, "internal/usecase/sign_up.go"),
		"internal/service/billing/charge.go": readFile(t, dir, "internal/service/billing/charge.go"),
	}
	for path, content := range handMade {
		keep[path] = content
	}

	// goro.yaml: drop the SignUp method and the Billing service.
	cfg := newConfig(t, dir)
	cfg.UseCase.Methods = []string{"SignIn"}
	cfg.UseCase.Deps = []string{"Users"}
	cfg.Services = []config.Service{cfg.Services[0]}

	report := update(t, cfg)

	t.Run("nothing is deleted", func(t *testing.T) {
		for path, content := range keep {
			if !exists(dir, path) {
				t.Errorf("%s was deleted", path)
				continue
			}
			if readFile(t, dir, path) != content {
				t.Errorf("%s was changed", path)
			}
		}
	})

	t.Run("unlisted code is reported as orphaned", func(t *testing.T) {
		want := []string{
			"internal/adapter/mysqlrepo/legacyrepo",
			"internal/service/billing",
			"internal/service/mailer",
			"internal/usecase/dep_billing.go",
			"internal/usecase/sign_up.go",
		}
		if got := orphans(report); !reflect.DeepEqual(got, want) {
			t.Errorf("orphans:\n got %q\nwant %q", got, want)
		}
	})

	t.Run("generated wiring stops using orphans", func(t *testing.T) {
		if container := readFile(t, dir, "internal/app/container.go"); strings.Contains(container, "billing") {
			t.Errorf("container.go still wires the removed Billing service")
		}
	})
}

func TestUpdateRegeneratesMarkedFiles(t *testing.T) {
	t.Run("adding a storage updates app.go and database.go", func(t *testing.T) {
		dir := t.TempDir()
		initProject(t, newConfig(t, dir))

		// goro.yaml: add pgsqlx to storages.
		cfg := newConfig(t, dir)
		cfg.Storages = append(cfg.Storages, "pgsqlx")
		report := update(t, cfg)

		if app := readFile(t, dir, "internal/app/app.go"); !strings.Contains(app, "pgsqlx") {
			t.Errorf("app.go doesn't define the pgsqlx connection")
		}
		if db := readFile(t, dir, "internal/app/database.go"); !strings.Contains(db, "newPgSqlxConnect") {
			t.Errorf("database.go doesn't build the pgsqlx connection")
		}
		assertAction(t, report, "internal/app/app.go", commands.ActionOverwrite)
		assertAction(t, report, "internal/app/database.go", commands.ActionOverwrite)
		assertMarkedFilesMatchFreshInit(t, dir, cfg)
	})

	t.Run("after adding a service dependency service.go and container.go agree", func(t *testing.T) {
		dir := t.TempDir()
		initProject(t, newConfig(t, dir))

		// goro.yaml: add the AuditRepo adapter as a dependency of Users.
		cfg := newConfig(t, dir)
		addAuditRepo(&cfg)
		report := update(t, cfg)

		params := funcParams(parseGo(t, dir, "internal/service/users/service.go"), "NewService")
		args := callArgs(parseGo(t, dir, "internal/app/container.go"), "users", "NewService")
		if params != 2 || args != 2 {
			t.Errorf("users.NewService takes %d parameters and container.go passes %d arguments, want 2 and 2", params, args)
		}
		assertAction(t, report, "internal/service/users/service.go", commands.ActionOverwrite)
		assertAction(t, report, "internal/app/container.go", commands.ActionOverwrite)
		assertMarkedFilesMatchFreshInit(t, dir, cfg)
	})
}

func TestUpdateSkipsUnmarkedFileAtGeneratedPath(t *testing.T) {
	t.Run("container.go without the marker", func(t *testing.T) {
		dir := t.TempDir()
		initProject(t, newConfig(t, dir))

		const path = "internal/app/container.go"
		owned := handEdit(removeMarker(readFile(t, dir, path)))
		writeFile(t, dir, path, owned)

		// goro.yaml: a change that the generated container.go would follow.
		cfg := newConfig(t, dir)
		addAuditRepo(&cfg)
		report := update(t, cfg)

		if readFile(t, dir, path) != owned {
			t.Errorf("update wrote %s, which has no marker", path)
		}
		assertAction(t, report, path, commands.ActionSkipUnmarked)
	})

	t.Run("service.go without the marker gets no new dep files", func(t *testing.T) {
		dir := t.TempDir()
		initProject(t, newConfig(t, dir))

		const path = "internal/service/users/service.go"
		owned := handEdit(removeMarker(readFile(t, dir, path)))
		writeFile(t, dir, path, owned)

		cfg := newConfig(t, dir)
		addAuditRepo(&cfg)
		report := update(t, cfg)

		if readFile(t, dir, path) != owned {
			t.Errorf("update wrote %s, which has no marker", path)
		}
		if exists(dir, "internal/service/users/dep_audit_repo.go") {
			t.Errorf("update created dep_audit_repo.go next to a user-owned service.go, which would declare the interface twice")
		}
		assertAction(t, report, path, commands.ActionSkipUnmarked)
	})
}

func TestDependencyInterfacesLiveInDepFiles(t *testing.T) {
	t.Run("init creates one user-owned dep file per dependency", func(t *testing.T) {
		dir := t.TempDir()
		initProject(t, newConfig(t, dir))

		deps := []struct {
			path, iface string
			transactor  bool
		}{
			{path: "internal/usecase/dep_users.go", iface: "users"},
			{path: "internal/usecase/dep_billing.go", iface: "billing"},
			{path: "internal/service/users/dep_user_repo.go", iface: "userRepo", transactor: true},
		}
		for _, d := range deps {
			if !exists(dir, d.path) {
				t.Errorf("init didn't create %s", d.path)
				continue
			}
			if hasMarker(readFile(t, dir, d.path)) {
				t.Errorf("%s carries the generated marker, but dep files belong to the user", d.path)
			}
			it, ok := interfaces(parseGo(t, dir, d.path))[d.iface]
			if !ok {
				t.Errorf("%s doesn't declare interface %s", d.path, d.iface)
				continue
			}
			if d.transactor && !embeds(it, "service", "Transactor") {
				t.Errorf("%s: interface %s doesn't embed service.Transactor for a transactional adapter", d.path, d.iface)
			}
		}

		for _, path := range []string{"internal/usecase/usecase.go", "internal/service/users/service.go"} {
			f := parseGo(t, dir, path)
			if !hasMarker(readFile(t, dir, path)) {
				t.Errorf("%s has no generated marker", path)
			}
			for name := range interfaces(f) {
				t.Errorf("generated %s declares dependency interface %s; it belongs in a dep file", path, name)
			}
		}
	})

	t.Run("an edited interface survives update", func(t *testing.T) {
		dir := t.TempDir()
		initProject(t, newConfig(t, dir))

		edited := map[string]string{
			"internal/usecase/dep_users.go": "package usecase\n\n" +
				"type users interface {\n\tCreate() error\n}\n",
			"internal/service/users/dep_user_repo.go": "package users\n\n" +
				"import (\n\t\"context\"\n\n\t\"demo/internal/service\"\n)\n\n" +
				"type userRepo interface {\n\tservice.Transactor\n\tGetByID(ctx context.Context, id int64) (string, error)\n}\n",
		}
		for path, content := range edited {
			writeFile(t, dir, path, content)
		}

		report := update(t, newConfig(t, dir))

		for path, content := range edited {
			if !exists(dir, path) {
				t.Errorf("%s was deleted", path)
				continue
			}
			if readFile(t, dir, path) != content {
				t.Errorf("update changed the edited interface in %s", path)
			}
			assertAction(t, report, path, commands.ActionSkip)
		}
	})

	t.Run("adding a dependency creates its file", func(t *testing.T) {
		dir := t.TempDir()
		initProject(t, newConfig(t, dir))

		// goro.yaml: Users depends on a new AuditRepo adapter, and the use case on a new
		// Notifier service.
		cfg := newConfig(t, dir)
		addAuditRepo(&cfg)
		cfg.Services = append(cfg.Services, config.Service{Name: "Notifier", Methods: []string{"Send"}})
		cfg.UseCase.Deps = append(cfg.UseCase.Deps, "Notifier")
		report := update(t, cfg)

		created := []struct {
			path, iface string
			transactor  bool
		}{
			{path: "internal/service/users/dep_audit_repo.go", iface: "auditRepo", transactor: true},
			{path: "internal/usecase/dep_notifier.go", iface: "notifier"},
		}
		for _, c := range created {
			if !exists(dir, c.path) {
				t.Errorf("update didn't create %s", c.path)
				continue
			}
			it, ok := interfaces(parseGo(t, dir, c.path))[c.iface]
			if !ok {
				t.Errorf("%s doesn't declare interface %s", c.path, c.iface)
				continue
			}
			if c.transactor && !embeds(it, "service", "Transactor") {
				t.Errorf("%s: interface %s doesn't embed service.Transactor for a transactional adapter", c.path, c.iface)
			}
			assertAction(t, report, c.path, commands.ActionCreate)
		}
	})
}

func TestUpdateDryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	initProject(t, newConfig(t, dir))
	writeFile(t, dir, "internal/service/mailer/mailer.go", "package mailer\n\nfunc Send() error { return nil }\n")
	writeFile(t, dir, "internal/usecase/sign_in.go", handEdit(readFile(t, dir, "internal/usecase/sign_in.go")))
	ageFiles(t, dir)
	before := snapshot(t, dir)

	// goro.yaml: add pgsqlx to storages and a SignOut use case method.
	cfg := newConfig(t, dir)
	cfg.Storages = append(cfg.Storages, "pgsqlx")
	cfg.UseCase.Methods = append(cfg.UseCase.Methods, "SignOut")

	var out bytes.Buffer
	dry, err := commands.Update(cfg, commands.Options{SkipGoMod: true, DryRun: true, Out: &out})
	if err != nil {
		t.Fatalf("update --dry-run: %v", err)
	}

	t.Run("nothing is written", func(t *testing.T) {
		assertNoChanges(t, before, snapshot(t, dir))
	})

	t.Run("each file gets an action", func(t *testing.T) {
		want := []commands.FileAction{
			{Path: "internal/usecase/sign_out.go", Action: commands.ActionCreate},
			{Path: "internal/app/app.go", Action: commands.ActionOverwrite},
			{Path: "internal/app/logger.go", Action: commands.ActionUnchanged},
			{Path: "internal/usecase/sign_in.go", Action: commands.ActionSkip},
			{Path: "internal/service/mailer", Action: commands.ActionOrphan},
		}
		for _, w := range want {
			assertAction(t, dry, w.Path, w.Action)
		}
	})

	t.Run("the actions are printed", func(t *testing.T) {
		printed := out.String()
		for _, f := range dry.Files {
			if f.Action == commands.ActionUnchanged {
				continue
			}
			if !hasLineWith(printed, string(f.Action), f.Path) {
				t.Errorf("output has no line with %q and %s", f.Action, f.Path)
			}
		}
		for _, a := range []commands.Action{commands.ActionCreate, commands.ActionOverwrite, commands.ActionSkip, commands.ActionOrphan} {
			if !strings.Contains(printed, string(a)) {
				t.Errorf("output shows no %q action:\n%s", a, printed)
			}
		}
	})

	t.Run("the real update does what the dry run reported", func(t *testing.T) {
		applied := update(t, cfg)
		if got, want := sortedActions(applied), sortedActions(dry); !reflect.DeepEqual(got, want) {
			t.Errorf("update actions differ from the dry run:\n got %v\nwant %v", got, want)
		}
	})
}

func TestSecondUpdateChangesNothing(t *testing.T) {
	dir := t.TempDir()
	initProject(t, newConfig(t, dir))

	// goro.yaml: a storage, a use case method and a service dependency added.
	cfg := newConfig(t, dir)
	cfg.Storages = append(cfg.Storages, "pgsqlx")
	cfg.UseCase.Methods = append(cfg.UseCase.Methods, "SignOut")
	addAuditRepo(&cfg)
	update(t, cfg)

	ageFiles(t, dir)
	before := snapshot(t, dir)
	report := update(t, cfg)

	assertNoChanges(t, before, snapshot(t, dir))
	for _, f := range report.Files {
		if f.Action == commands.ActionCreate || f.Action == commands.ActionOverwrite {
			t.Errorf("second update reports %q for %s", f.Action, f.Path)
		}
	}
	assertAction(t, report, "internal/app/container.go", commands.ActionUnchanged)
	assertAction(t, report, "internal/usecase/sign_in.go", commands.ActionSkip)
}

func hasLineWith(text string, parts ...string) bool {
	for _, line := range strings.Split(text, "\n") {
		all := true
		for _, p := range parts {
			if !strings.Contains(line, p) {
				all = false
				break
			}
		}
		if all {
			return true
		}
	}

	return false
}

func sortedActions(r commands.Report) []string {
	res := make([]string, 0, len(r.Files))
	for _, f := range r.Files {
		res = append(res, string(f.Action)+" "+f.Path)
	}
	sort.Strings(res)

	return res
}
