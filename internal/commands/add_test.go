package commands_test

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"

	"github.com/hanagantig/goro/internal/commands"
	"github.com/hanagantig/goro/internal/config"
)

// Rules under test: the goro add commands (ADR 0002). Each one validates, edits goro.yaml in the
// project root in place, inserting lines and changing no other line, and runs update.

// addOpts runs add without the go mod steps and without output.
var addOpts = commands.Options{SkipGoMod: true, Out: io.Discard}

// addProject makes a project as ADR 0002 has it: testdata/add/goro.yaml in the project root, and
// the code init generates from it. It returns the project dir and the path of its goro.yaml.
func addProject(t *testing.T) (dir, configPath string) {
	t.Helper()

	dir = t.TempDir()
	configPath = filepath.Join(dir, "goro.yaml")
	writeFile(t, dir, "goro.yaml", readTestdata(t, "add/goro.yaml"))
	initProject(t, loadProjectConfig(t, configPath))

	return dir, configPath
}

// loadProjectConfig loads the goro.yaml at path, which sits in the project root. It sets the work
// dir to the file's directory itself, so the add tests don't depend on how LoadConfig resolves
// work_dir: . (config_test.go covers that).
func loadProjectConfig(t *testing.T, path string) config.Config {
	t.Helper()

	cfg, err := commands.LoadConfig(path)
	if err != nil {
		t.Fatalf("load %s: %v", path, err)
	}
	cfg.App.WorkDir = filepath.Dir(path)

	return cfg
}

func readTestdata(t *testing.T, rel string) string {
	t.Helper()

	b, err := os.ReadFile(filepath.Join("testdata", filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}

func findService(cfg config.Config, name string) (config.Service, bool) {
	for _, s := range cfg.Services {
		if s.Name == name {
			return s, true
		}
	}

	return config.Service{}, false
}

func findAdapter(cfg config.Config, name string) (config.Adapter, bool) {
	for _, a := range cfg.Adapters {
		if a.Name == name {
			return a, true
		}
	}

	return config.Adapter{}, false
}

func assertStrings(t *testing.T, what string, got, want []string) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s = %q, want %q", what, got, want)
	}
}

// assertCreated checks that the report creates each path and that the file is there.
func assertCreated(t *testing.T, r commands.Report, dir string, paths ...string) {
	t.Helper()

	for _, path := range paths {
		assertAction(t, r, path, commands.ActionCreate)
		if !exists(dir, path) {
			t.Errorf("%s wasn't created", path)
		}
	}
}

func TestAddUseCase(t *testing.T) {
	dir, path := addProject(t)

	report, err := commands.AddUseCase(path, "SignOut", addOpts)
	if err != nil {
		t.Fatalf("add usecase SignOut: %v", err)
	}

	cfg := loadProjectConfig(t, path)
	assertStrings(t, "use_case.methods in goro.yaml", cfg.UseCase.Methods, []string{"SignIn", "Pong", "SignOut"})
	assertOnlyInsertedLines(t, path)
	assertCreated(t, report, dir, "internal/usecase/sign_out.go")
	assertMatchesFreshInit(t, dir, cfg)
}

func TestAddService(t *testing.T) {
	dir, path := addProject(t)

	// goro add service Billing --deps UserRepo,Users --methods Charge,Refund
	svc := config.Service{Name: "Billing", Deps: []string{"UserRepo", "Users"}, Methods: []string{"Charge", "Refund"}}
	report, err := commands.AddService(path, svc, addOpts)
	if err != nil {
		t.Fatalf("add service Billing: %v", err)
	}

	cfg := loadProjectConfig(t, path)
	got, ok := findService(cfg, "Billing")
	if !ok {
		t.Fatalf("goro.yaml lists no Billing service")
	}
	assertStrings(t, "Billing methods in goro.yaml", got.Methods, svc.Methods)
	assertStrings(t, "Billing deps in goro.yaml", got.Deps, svc.Deps)
	assertOnlyInsertedLines(t, path)

	assertCreated(t, report, dir,
		"internal/service/billing/service.go",
		"internal/service/billing/charge.go",
		"internal/service/billing/refund.go",
		"internal/service/billing/dep_user_repo.go",
		"internal/service/billing/dep_users.go",
	)
	if exists(dir, "internal/service/billing/dep_user_repo.go") {
		it, ok := interfaces(parseGo(t, dir, "internal/service/billing/dep_user_repo.go"))["userRepo"]
		if !ok || !embeds(it, "service", "Transactor") {
			t.Errorf("dep_user_repo.go doesn't declare userRepo embedding service.Transactor for the mysql adapter")
		}
	}

	t.Run("container.go wires it", func(t *testing.T) {
		assertAction(t, report, "internal/app/container.go", commands.ActionOverwrite)
		if !exists(dir, "internal/service/billing/service.go") {
			t.Fatalf("no service.go for Billing")
		}
		params := funcParams(parseGo(t, dir, "internal/service/billing/service.go"), "NewService")
		args := callArgs(parseGo(t, dir, "internal/app/container.go"), "billing", "NewService")
		if params != 2 || args != 2 {
			t.Errorf("billing.NewService takes %d parameters and container.go passes %d arguments, want 2 and 2", params, args)
		}
	})

	assertMatchesFreshInit(t, dir, cfg)
}

func TestAddMethod(t *testing.T) {
	t.Run("to a service", func(t *testing.T) {
		dir, path := addProject(t)

		report, err := commands.AddMethod(path, "Users.Delete", addOpts)
		if err != nil {
			t.Fatalf("add method Users.Delete: %v", err)
		}

		cfg := loadProjectConfig(t, path)
		svc, _ := findService(cfg, "Users")
		assertStrings(t, "Users methods in goro.yaml", svc.Methods, []string{"Create", "Delete"})
		assertOnlyInsertedLines(t, path)
		assertCreated(t, report, dir, "internal/service/users/delete.go")
		assertMatchesFreshInit(t, dir, cfg)
	})

	t.Run("to an adapter", func(t *testing.T) {
		dir, path := addProject(t)

		report, err := commands.AddMethod(path, "UserRepo.Save", addOpts)
		if err != nil {
			t.Fatalf("add method UserRepo.Save: %v", err)
		}

		cfg := loadProjectConfig(t, path)
		adapter, _ := findAdapter(cfg, "UserRepo")
		assertStrings(t, "UserRepo methods in goro.yaml", adapter.Methods, []string{"GetByID", "Save"})
		assertOnlyInsertedLines(t, path)
		assertCreated(t, report, dir, "internal/adapter/mysqlrepo/userrepo/save.go")
		assertMatchesFreshInit(t, dir, cfg)
	})
}

func TestAddAdapter(t *testing.T) {
	t.Run("a new storage is added to storages", func(t *testing.T) {
		dir, path := addProject(t)

		// goro add adapter PaymentsRepo --storage pgsqlx --methods Save
		adapter := config.Adapter{Name: "PaymentsRepo", Storage: "pgsqlx", Methods: []string{"Save"}}
		report, err := commands.AddAdapter(path, adapter, addOpts)
		if err != nil {
			t.Fatalf("add adapter PaymentsRepo: %v", err)
		}

		cfg := loadProjectConfig(t, path)
		storages := make([]string, 0, len(cfg.Storages))
		for _, s := range cfg.Storages {
			storages = append(storages, s.String())
		}
		assertStrings(t, "storages in goro.yaml", storages, []string{"mysql", "pgsqlx"})
		got, ok := findAdapter(cfg, "PaymentsRepo")
		if !ok {
			t.Fatalf("goro.yaml lists no PaymentsRepo adapter")
		}
		if got.Storage != "pgsqlx" {
			t.Errorf("PaymentsRepo storage in goro.yaml = %q, want pgsqlx", got.Storage)
		}
		assertStrings(t, "PaymentsRepo methods in goro.yaml", got.Methods, adapter.Methods)
		assertOnlyInsertedLines(t, path)

		assertCreated(t, report, dir,
			"internal/adapter/pgsqlxrepo/transactor.go",
			"internal/adapter/pgsqlxrepo/paymentsrepo/repository.go",
			"internal/adapter/pgsqlxrepo/paymentsrepo/save.go",
		)
		assertAction(t, report, "internal/app/app.go", commands.ActionOverwrite)
		assertAction(t, report, "internal/app/database.go", commands.ActionOverwrite)
		assertMatchesFreshInit(t, dir, cfg)
	})

	t.Run("a listed storage isn't added again", func(t *testing.T) {
		dir, path := addProject(t)

		report, err := commands.AddAdapter(path, config.Adapter{Name: "AuditRepo", Storage: "mysql", Methods: []string{"Write"}}, addOpts)
		if err != nil {
			t.Fatalf("add adapter AuditRepo: %v", err)
		}

		cfg := loadProjectConfig(t, path)
		if len(cfg.Storages) != 1 || cfg.Storages[0] != "mysql" {
			t.Errorf("storages in goro.yaml = %q, want [mysql]", cfg.Storages)
		}
		if _, ok := findAdapter(cfg, "AuditRepo"); !ok {
			t.Errorf("goro.yaml lists no AuditRepo adapter")
		}
		assertOnlyInsertedLines(t, path)
		assertCreated(t, report, dir,
			"internal/adapter/mysqlrepo/auditrepo/repository.go",
			"internal/adapter/mysqlrepo/auditrepo/write.go",
		)
		assertMatchesFreshInit(t, dir, cfg)
	})
}

// TestAddFailuresChangeNothing: an add that fails validation leaves goro.yaml and the code as they
// are, and its error names the value at fault.
func TestAddFailuresChangeNothing(t *testing.T) {
	cases := []struct {
		name string
		// setup, when set, changes the project before the add.
		setup func(t *testing.T, dir string)
		add   func(path string) (commands.Report, error)
		// inErr lists what the error has to name.
		inErr []string
	}{
		{
			name:  "add usecase with a method goro.yaml already lists",
			add:   func(p string) (commands.Report, error) { return commands.AddUseCase(p, "SignIn", addOpts) },
			inErr: []string{"SignIn"},
		},
		{
			name:  "add usecase with a bad method name",
			add:   func(p string) (commands.Report, error) { return commands.AddUseCase(p, "signOut", addOpts) },
			inErr: []string{"signOut"},
		},
		{
			// DepUsers would get the file dep_users.go, the dep file of the Users dependency.
			name:  "add usecase with a method whose file is a dep file",
			add:   func(p string) (commands.Report, error) { return commands.AddUseCase(p, "DepUsers", addOpts) },
			inErr: []string{"DepUsers"},
		},
		{
			// Usecase would get the file usecase.go, which goro owns.
			name:  "add usecase with a method whose file is goro's",
			add:   func(p string) (commands.Report, error) { return commands.AddUseCase(p, "Usecase", addOpts) },
			inErr: []string{"Usecase"},
		},
		{
			name: "add service with a name goro.yaml already lists",
			add: func(p string) (commands.Report, error) {
				return commands.AddService(p, config.Service{Name: "Users"}, addOpts)
			},
			inErr: []string{"Users"},
		},
		{
			name: "add service with the name of an adapter",
			add: func(p string) (commands.Report, error) {
				return commands.AddService(p, config.Service{Name: "UserRepo"}, addOpts)
			},
			inErr: []string{"UserRepo"},
		},
		{
			name: "add service with a bad name",
			add: func(p string) (commands.Report, error) {
				return commands.AddService(p, config.Service{Name: "billing"}, addOpts)
			},
			inErr: []string{"billing"},
		},
		{
			name: "add service with an unknown dependency",
			add: func(p string) (commands.Report, error) {
				return commands.AddService(p, config.Service{Name: "Billing", Deps: []string{"PaymentsRepo"}}, addOpts)
			},
			inErr: []string{"PaymentsRepo"},
		},
		{
			name: "add service with a bad method name",
			add: func(p string) (commands.Report, error) {
				return commands.AddService(p, config.Service{Name: "Billing", Methods: []string{"charge"}}, addOpts)
			},
			inErr: []string{"charge"},
		},
		{
			name: "add adapter with a name goro.yaml already lists",
			add: func(p string) (commands.Report, error) {
				return commands.AddAdapter(p, config.Adapter{Name: "UserRepo", Storage: "mysql"}, addOpts)
			},
			inErr: []string{"UserRepo"},
		},
		{
			name: "add adapter with the name of a service",
			add: func(p string) (commands.Report, error) {
				return commands.AddAdapter(p, config.Adapter{Name: "Users", Storage: "mysql"}, addOpts)
			},
			inErr: []string{"Users"},
		},
		{
			name: "add adapter with a bad name",
			add: func(p string) (commands.Report, error) {
				return commands.AddAdapter(p, config.Adapter{Name: "Audit_Repo", Storage: "mysql"}, addOpts)
			},
			inErr: []string{"Audit_Repo"},
		},
		{
			name: "add adapter with an unsupported storage",
			add: func(p string) (commands.Report, error) {
				return commands.AddAdapter(p, config.Adapter{Name: "PaymentsRepo", Storage: "mongo"}, addOpts)
			},
			inErr: []string{"mongo"},
		},
		{
			name: "add adapter with a bad method name",
			add: func(p string) (commands.Report, error) {
				return commands.AddAdapter(p, config.Adapter{Name: "PaymentsRepo", Storage: "pgsqlx", Methods: []string{"save"}}, addOpts)
			},
			inErr: []string{"save"},
		},
		{
			name:  "add method that goro.yaml already lists",
			add:   func(p string) (commands.Report, error) { return commands.AddMethod(p, "Users.Create", addOpts) },
			inErr: []string{"Create"},
		},
		{
			name:  "add method with a bad method name",
			add:   func(p string) (commands.Report, error) { return commands.AddMethod(p, "Users.delete", addOpts) },
			inErr: []string{"delete"},
		},
		{
			// DepUserRepo would get the file dep_user_repo.go, the dep file of UserRepo.
			name:  "add method whose file is a dep file",
			add:   func(p string) (commands.Report, error) { return commands.AddMethod(p, "Users.DepUserRepo", addOpts) },
			inErr: []string{"DepUserRepo"},
		},
		{
			// Service would get the file service.go, which goro owns.
			name:  "add method to a service whose file is goro's",
			add:   func(p string) (commands.Report, error) { return commands.AddMethod(p, "Users.Service", addOpts) },
			inErr: []string{"Service"},
		},
		{
			// Repository would get the file repository.go, which goro owns.
			name:  "add method to an adapter whose file is goro's",
			add:   func(p string) (commands.Report, error) { return commands.AddMethod(p, "UserRepo.Repository", addOpts) },
			inErr: []string{"Repository"},
		},
		{
			// Transactor would get transactor.go, the name of goro's transactor files, and
			// clash with the Transactor that a transactional adapter's Repository embeds.
			name:  "add method to an adapter named like its transactor",
			add:   func(p string) (commands.Report, error) { return commands.AddMethod(p, "UserRepo.Transactor", addOpts) },
			inErr: []string{"Transactor"},
		},
		{
			name:  "add method to a name that is neither a service nor an adapter",
			add:   func(p string) (commands.Report, error) { return commands.AddMethod(p, "Payments.Save", addOpts) },
			inErr: []string{"Payments"},
		},
		{
			name: "add method to a name that is both a service and an adapter",
			setup: func(t *testing.T, dir string) {
				// A hand edit: an adapter named like the Users service, appended to adapters,
				// the last list in the file.
				appendFile(t, dir, "goro.yaml", "  - name: Users\n    storage: mysql\n")
			},
			add:   func(p string) (commands.Report, error) { return commands.AddMethod(p, "Users.Delete", addOpts) },
			inErr: []string{"Users"},
		},
		// goro add only inserts lines, and a flow list ([a, b]) can't grow by a line, so add
		// refuses to rewrite one and names the file and the list instead.
		{
			name: "add usecase when use_case.methods is a flow list",
			setup: func(t *testing.T, dir string) {
				replaceInFile(t, dir, "goro.yaml",
					"  methods:\n    - SignIn\n    - Pong # the /ping handler that init generates calls it\n",
					"  methods: [SignIn, Pong]\n")
			},
			add:   func(p string) (commands.Report, error) { return commands.AddUseCase(p, "SignOut", addOpts) },
			inErr: []string{"goro.yaml", "use_case", "methods"},
		},
		{
			name: "add adapter with a new storage when storages is a flow list",
			setup: func(t *testing.T, dir string) {
				replaceInFile(t, dir, "goro.yaml", "storages:\n  - mysql\n  # - pgsqlx\n", "storages: [mysql]\n")
			},
			add: func(p string) (commands.Report, error) {
				return commands.AddAdapter(p, config.Adapter{Name: "PaymentsRepo", Storage: "pgsqlx"}, addOpts)
			},
			inErr: []string{"goro.yaml", "storages"},
		},
		{
			name: "add method when the adapter's methods is a flow list",
			setup: func(t *testing.T, dir string) {
				replaceInFile(t, dir, "goro.yaml", "    methods:\n      - GetByID\n", "    methods: [GetByID]\n")
			},
			add:   func(p string) (commands.Report, error) { return commands.AddMethod(p, "UserRepo.Save", addOpts) },
			inErr: []string{"goro.yaml", "UserRepo", "methods"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir, path := addProject(t)
			if c.setup != nil {
				c.setup(t, dir)
			}
			ageFiles(t, dir)
			before := snapshot(t, dir)

			_, err := c.add(path)
			if err == nil {
				t.Errorf("the add succeeded, want an error")
			} else {
				for _, want := range c.inErr {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("error %q doesn't name %q", err, want)
					}
				}
			}
			assertNoChanges(t, before, snapshot(t, dir))
		})
	}
}

// replaceInFile replaces old, which must be there, with new in a file of dir.
func replaceInFile(t *testing.T, dir, rel, old, new string) {
	t.Helper()

	content := readFile(t, dir, rel)
	if !strings.Contains(content, old) {
		t.Fatalf("%s doesn't contain %q", rel, old)
	}
	writeFile(t, dir, rel, strings.Replace(content, old, new, 1))
}

func appendFile(t *testing.T, dir, rel, content string) {
	t.Helper()

	writeFile(t, dir, rel, readFile(t, dir, rel)+content)
}

// addStep is one goro add command, run against the goro.yaml at path.
type addStep struct {
	name string
	run  func(path string, opts commands.Options) (commands.Report, error)
}

// addSteps grows the project of testdata/add/goro.yaml with every goro add subcommand, in the
// order a user would run them. testdata/add/goro.golden.yaml is goro.yaml after them.
var addSteps = []addStep{
	{
		name: "add adapter PaymentsRepo --storage pgsqlx --methods Save",
		run: func(path string, opts commands.Options) (commands.Report, error) {
			return commands.AddAdapter(path, config.Adapter{Name: "PaymentsRepo", Storage: "pgsqlx", Methods: []string{"Save"}}, opts)
		},
	},
	{
		name: "add service Billing --deps PaymentsRepo --methods Charge,Refund",
		run: func(path string, opts commands.Options) (commands.Report, error) {
			return commands.AddService(path, config.Service{Name: "Billing", Deps: []string{"PaymentsRepo"}, Methods: []string{"Charge", "Refund"}}, opts)
		},
	},
	{
		name: "add method Billing.Void",
		run: func(path string, opts commands.Options) (commands.Report, error) {
			return commands.AddMethod(path, "Billing.Void", opts)
		},
	},
	{
		name: "add method UserRepo.Save",
		run: func(path string, opts commands.Options) (commands.Report, error) {
			return commands.AddMethod(path, "UserRepo.Save", opts)
		},
	},
	{
		name: "add usecase SignOut",
		run: func(path string, opts commands.Options) (commands.Report, error) {
			return commands.AddUseCase(path, "SignOut", opts)
		},
	},
}

func runAddSteps(t *testing.T, path string, opts commands.Options) {
	t.Helper()

	for _, s := range addSteps {
		if _, err := s.run(path, opts); err != nil {
			t.Fatalf("%s: %v", s.name, err)
		}
	}
}

// assertGolden compares the file at path with testdata/<golden>.
func assertGolden(t *testing.T, path, golden string) {
	t.Helper()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got, want := string(b), readTestdata(t, golden)
	if got == want {
		return
	}

	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(want),
		B:        difflib.SplitLines(got),
		FromFile: "testdata/" + golden,
		ToFile:   path,
		Context:  3,
	})
	t.Errorf("%s differs from testdata/%s:\n%s", path, golden, diff)
}

// lines splits text into lines that keep their line endings, so joining them gives text back
// byte for byte.
func lines(text string) []string {
	res := strings.SplitAfter(text, "\n")
	if res[len(res)-1] == "" {
		res = res[:len(res)-1]
	}

	return res
}

// assertOnlyInsertedLines checks that goro.yaml at path is testdata/add/goro.yaml with lines
// inserted: dropping the added lines gives the original back byte for byte, so no line of it
// moved or changed (ADR 0002, consequences observed).
func assertOnlyInsertedLines(t *testing.T, path string) {
	t.Helper()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	original, edited := lines(readTestdata(t, "add/goro.yaml")), lines(string(b))

	// The original has to be a subsequence of the edited file; matching greedily finds one when
	// there is one.
	i := 0
	for _, line := range edited {
		if i < len(original) && line == original[i] {
			i++
		}
	}
	if i < len(original) {
		t.Errorf("%s isn't testdata/add/goro.yaml with lines inserted: line %d of the original, %q, changed, moved or went missing:\n%s",
			path, i+1, original[i], b)
	}
}

func TestAddEditsGoroYamlInPlace(t *testing.T) {
	dir, path := addProject(t)

	runAddSteps(t, path, addOpts)

	// New list items go after the last item of the list, before trailing commented-out items,
	// at the list's indentation; nothing else in the file changes.
	assertGolden(t, path, "add/goro.golden.yaml")
	assertOnlyInsertedLines(t, path)
	assertMatchesFreshInit(t, dir, loadProjectConfig(t, path))
}
