package commands_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/hanagantig/goro/internal/commands"
	"github.com/hanagantig/goro/internal/config"
)

// namesValue reports whether msg names v the way an error should: quoted ('v' or "v") or
// qualified by what it belongs to (Users.Service). A bare word could match any other part of the
// message.
func namesValue(msg, v string) bool {
	q := regexp.QuoteMeta(v)

	return regexp.MustCompile(`'` + q + `'|"` + q + `"|\w\.` + q + `(?:[^\w]|$)`).MatchString(msg)
}

// errorCheck is what a failing call's error has to hold.
type errorCheck struct {
	// names lists values the error names, quoted or qualified (see namesValue).
	names []string
	// mentions lists other text the error holds, such as the path of a list in goro.yaml.
	mentions []string
	// file is a path the error holds, such as the goro.yaml at fault.
	file string
}

func assertError(t *testing.T, err error, want errorCheck) {
	t.Helper()

	if err == nil {
		t.Errorf("the call succeeded, want an error")
		return
	}
	msg := err.Error()
	for _, v := range want.names {
		if !namesValue(msg, v) {
			t.Errorf("error %q doesn't name %q, quoted or qualified", msg, v)
		}
	}
	for _, m := range want.mentions {
		if !strings.Contains(msg, m) {
			t.Errorf("error %q doesn't mention %s", msg, m)
		}
	}
	if want.file != "" && !strings.Contains(msg, want.file) {
		t.Errorf("error %q doesn't name the file %s", msg, want.file)
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

func addsUseCase(method string) func(string) (commands.Report, error) {
	return func(p string) (commands.Report, error) { return commands.AddUseCase(p, method, addOpts) }
}

func addsService(svc config.Service) func(string) (commands.Report, error) {
	return func(p string) (commands.Report, error) { return commands.AddService(p, svc, addOpts) }
}

func addsAdapter(a config.Adapter) func(string) (commands.Report, error) {
	return func(p string) (commands.Report, error) { return commands.AddAdapter(p, a, addOpts) }
}

func addsMethod(ref string) func(string) (commands.Report, error) {
	return func(p string) (commands.Report, error) { return commands.AddMethod(p, ref, addOpts) }
}

// TestAddFailuresChangeNothing: an add that fails validation leaves goro.yaml and the code as they
// are, and its error names the value at fault. The project is testdata/add/goro.yaml: the Users
// service depends on UserRepo, a mysql adapter, so transactional.
func TestAddFailuresChangeNothing(t *testing.T) {
	// Each add's error names value, quoted or qualified.
	cases := []struct {
		name  string
		add   func(path string) (commands.Report, error)
		value string
	}{
		// Names goro.yaml lists already, or that map to the same file or package.
		{"add usecase with a method goro.yaml already lists", addsUseCase("SignIn"), "SignIn"},
		{"add usecase with a method whose file is a listed one's", addsUseCase("SignIN"), "SignIN"},
		{"add service with a name goro.yaml already lists", addsService(config.Service{Name: "Users"}), "Users"},
		{"add service with the name of an adapter", addsService(config.Service{Name: "UserRepo"}), "UserRepo"},
		// container.go imports the UserRepo adapter's package as userrepo already.
		{"add service whose package is an adapter's", addsService(config.Service{Name: "Userrepo"}), "Userrepo"},
		{"add adapter with a name goro.yaml already lists", addsAdapter(mysqlAdapter("UserRepo")), "UserRepo"},
		{"add adapter with the name of a service", addsAdapter(mysqlAdapter("Users")), "Users"},
		// internal/adapter/mysqlrepo/userrepo is UserRepo's.
		{"add adapter whose package is a listed adapter's", addsAdapter(mysqlAdapter("Userrepo")), "Userrepo"},
		{"add method that goro.yaml already lists", addsMethod("Users.Create"), "Create"},

		// Bad names.
		{"add usecase with a bad method name", addsUseCase("signOut"), "signOut"},
		{"add service with a bad name", addsService(config.Service{Name: "billing"}), "billing"},
		{
			"add service with a bad method name",
			addsService(config.Service{Name: "Billing", Methods: []string{"charge"}}), "charge",
		},
		{"add adapter with a bad name", addsAdapter(mysqlAdapter("Audit_Repo")), "Audit_Repo"},
		{
			"add adapter with a bad method name",
			addsAdapter(config.Adapter{Name: "PaymentsRepo", Storage: "pgsqlx", Methods: []string{"save"}}), "save",
		},
		{"add method with a bad method name", addsMethod("Users.delete"), "delete"},

		// Method files that clash with other files: dep files, goro's files, and files Go
		// builds only for tests or for one OS or architecture.
		{"add usecase with a method whose file is a dep file", addsUseCase("DepUsers"), "DepUsers"},
		{"add method whose file is a dep file", addsMethod("Users.DepUserRepo"), "DepUserRepo"},
		{"add usecase with a method whose file is goro's", addsUseCase("Usecase"), "Usecase"},
		{"add method to a service whose file is goro's", addsMethod("Users.Service"), "Service"},
		{"add method to an adapter whose file is goro's", addsMethod("UserRepo.Repository"), "Repository"},
		{"add usecase with a method whose file is a test file", addsUseCase("SignTest"), "SignTest"},
		{"add method whose file builds only on one OS", addsMethod("Users.ReadLinux"), "ReadLinux"},
		{"add method whose file builds only on one architecture", addsMethod("UserRepo.FetchArm"), "FetchArm"},

		// A transactional adapter's Repository embeds a Transactor: the field, and the methods
		// it promotes.
		{"add method named like the adapter's Transactor", addsMethod("UserRepo.Transactor"), "Transactor"},
		{"add method that shadows Transactor.NewTxContext", addsMethod("UserRepo.NewTxContext"), "NewTxContext"},
		{"add method that shadows Transactor.InTransaction", addsMethod("UserRepo.InTransaction"), "InTransaction"},
		{"add method that shadows Transactor.GetConn", addsMethod("UserRepo.GetConn"), "GetConn"},
		{"add method that shadows Transactor.RunTransaction", addsMethod("UserRepo.RunTransaction"), "RunTransaction"},

		// Dependencies and storages.
		{
			"add service with an unknown dependency",
			addsService(config.Service{Name: "Billing", Deps: []string{"PaymentsRepo"}}), "PaymentsRepo",
		},
		{
			"add service that depends on itself",
			addsService(config.Service{Name: "Billing", Deps: []string{"Billing"}}), "Billing",
		},
		{
			"add adapter with an unsupported storage",
			addsAdapter(config.Adapter{Name: "PaymentsRepo", Storage: "mongo"}), "mongo",
		},

		// Name.Method refs.
		{"add method to a name that is neither a service nor an adapter", addsMethod("Payments.Save"), "Payments"},
		{"add method with a ref without a method", addsMethod("Users"), "Users"},
		{"add method with a ref with an empty method", addsMethod("Users."), "Users."},
		{"add method with a ref with an empty name", addsMethod(".Save"), ".Save"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertAddFails(t, nil, c.add, func(string) errorCheck { return errorCheck{names: []string{c.value}} })
		})
	}

	// Adds on a hand-edited goro.yaml, and errors that hold more than a value.
	edited := []struct {
		name string
		// setup changes the project before the add.
		setup func(t *testing.T, dir string)
		add   func(path string) (commands.Report, error)
		// err gives the check for the goro.yaml at path.
		err func(path string) errorCheck
	}{
		{
			name: "add adapter without a storage",
			add:  addsAdapter(config.Adapter{Name: "PaymentsRepo"}),
			err: func(string) errorCheck {
				return errorCheck{names: []string{"PaymentsRepo"}, mentions: []string{"storage"}}
			},
		},
		{
			name: "add method to a name that is both a service and an adapter",
			setup: func(t *testing.T, dir string) {
				// An adapter named like the Users service, appended to adapters, the last list
				// in the file.
				appendFile(t, dir, "goro.yaml", "  - name: Users\n    storage: mysql\n")
			},
			add: addsMethod("Users.Delete"),
			err: func(string) errorCheck { return errorCheck{names: []string{"Users"}} },
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
			add: addsUseCase("SignOut"),
			err: func(p string) errorCheck { return errorCheck{file: p, mentions: []string{"use_case.methods"}} },
		},
		{
			name: "add adapter with a new storage when storages is a flow list",
			setup: func(t *testing.T, dir string) {
				replaceInFile(t, dir, "goro.yaml", "storages:\n  - mysql\n  # - pgsqlx\n", "storages: [mysql]\n")
			},
			add: addsAdapter(config.Adapter{Name: "PaymentsRepo", Storage: "pgsqlx"}),
			err: func(p string) errorCheck { return errorCheck{file: p, mentions: []string{"storages"}} },
		},
		{
			name: "add method when the adapter's methods is a flow list",
			setup: func(t *testing.T, dir string) {
				replaceInFile(t, dir, "goro.yaml", "    methods:\n      - GetByID\n", "    methods: [GetByID]\n")
			},
			add: addsMethod("UserRepo.Save"),
			err: func(p string) errorCheck { return errorCheck{file: p, mentions: []string{"adapters.UserRepo.methods"}} },
		},
		{
			name: "add adapter when adapters is an empty flow list",
			setup: func(t *testing.T, dir string) {
				// Users depends on nothing now, so the hand-edited goro.yaml is valid.
				replaceInFile(t, dir, "goro.yaml", "    deps: [UserRepo]\n", "")
				replaceInFile(t, dir, "goro.yaml",
					"adapters:\n  - name: UserRepo # the users table\n    storage: mysql\n    methods:\n      - GetByID\n",
					"adapters: []\n")
			},
			add: addsAdapter(mysqlAdapter("PaymentsRepo")),
			err: func(p string) errorCheck { return errorCheck{file: p, mentions: []string{"adapters"}} },
		},

		// add never asks: what goro.yaml lacks is an error that says what to set.
		{
			name: "add usecase when goro.yaml has no app name",
			setup: func(t *testing.T, dir string) {
				replaceInFile(t, dir, "goro.yaml", "app:\n  name: demo\n", "app:\n")
			},
			add: addsUseCase("SignOut"),
			err: func(p string) errorCheck { return errorCheck{file: p, mentions: []string{"app.name"}} },
		},
		{
			name: "add usecase when goro.yaml has no work_dir",
			setup: func(t *testing.T, dir string) {
				replaceInFile(t, dir, "goro.yaml", "  work_dir: . # the project, relative to this file\n", "")
			},
			add: addsUseCase("SignOut"),
			err: func(p string) errorCheck { return errorCheck{file: p, mentions: []string{"app.work_dir"}} },
		},
	}
	for _, c := range edited {
		t.Run(c.name, func(t *testing.T) { assertAddFails(t, c.setup, c.add, c.err) })
	}
}

func mysqlAdapter(name string) config.Adapter {
	return config.Adapter{Name: name, Storage: "mysql"}
}

// assertAddFails runs add on a project made by addProject and changed by setup, when set, and
// checks that it fails with an error that passes check and that nothing changed.
func assertAddFails(t *testing.T, setup func(t *testing.T, dir string), add func(path string) (commands.Report, error),
	check func(path string) errorCheck) {
	t.Helper()

	dir, path := addProject(t)
	if setup != nil {
		setup(t, dir)
	}
	ageFiles(t, dir)
	before := snapshot(t, dir)

	_, err := add(path)
	assertError(t, err, check(path))
	assertNoChanges(t, before, snapshot(t, dir))
}
