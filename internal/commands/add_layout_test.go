package commands_test

import (
	"testing"

	"github.com/hanagantig/goro/internal/commands"
	"github.com/hanagantig/goro/internal/config"
)

// Rules under test: goro add inserts lines into any goro.yaml, whatever it lists and however it is
// laid out (ADR 0002, consequences observed). These cases don't pin the layout of the inserted
// lines; they check that the file still says what it said, plus the addition.

// growCase is a goro.yaml, the adds run on it, and the change they make to its config.
type growCase struct {
	name  string
	yaml  string
	steps []addStep
	// want changes the config parsed from yaml into the config expected after the steps.
	want func(cfg *config.Config)
}

func runGrowCase(t *testing.T, c growCase) {
	t.Helper()

	dir, path := projectFrom(t, c.yaml)
	want := loadProjectConfig(t, path)
	c.want(&want)

	for _, s := range c.steps {
		previous := readFile(t, dir, "goro.yaml")
		if _, err := s.run(path, addOpts); err != nil {
			t.Fatalf("%s: %v", s.name, err)
		}
		// Each add leaves a goro.yaml that loads and only grew.
		loadProjectConfig(t, path)
		assertOnlyInsertedLines(t, previous, path)
	}

	got := loadProjectConfig(t, path)
	assertConfigLists(t, got, want)
	assertOnlyInsertedLines(t, c.yaml, path)
	assertMatchesFreshInit(t, dir, got)
}

func step(name string, run func(path string) (commands.Report, error)) addStep {
	return addStep{name: name, run: func(path string, _ commands.Options) (commands.Report, error) { return run(path) }}
}

const minimalApp = `app:
  name: demo
  module: demo
  work_dir: .
`

func TestAddCreatesMissingKeys(t *testing.T) {
	cases := []growCase{
		{
			// goro add has to be able to extend its own output.
			name: "a service added without methods, then a method on it",
			yaml: readTestdata(t, "add/goro.yaml"),
			steps: []addStep{
				step("add service Mailer", addsService(config.Service{Name: "Mailer"})),
				step("add method Mailer.Send", addsMethod("Mailer.Send")),
			},
			want: func(cfg *config.Config) {
				cfg.Services = append(cfg.Services, config.Service{Name: "Mailer", Methods: []string{"Send"}})
			},
		},
		{
			name: "an adapter added without methods, then a method on it",
			yaml: readTestdata(t, "add/goro.yaml"),
			steps: []addStep{
				step("add adapter AuditRepo --storage mysql", addsAdapter(config.Adapter{Name: "AuditRepo", Storage: "mysql"})),
				step("add method AuditRepo.Write", addsMethod("AuditRepo.Write")),
			},
			want: func(cfg *config.Config) {
				cfg.Adapters = append(cfg.Adapters, config.Adapter{Name: "AuditRepo", Storage: "mysql", Methods: []string{"Write"}})
			},
		},
		{
			name: "no services key",
			yaml: minimalApp + "\nstorages:\n  - mysql\n\nadapters:\n  - name: UserRepo\n    storage: mysql\n    methods:\n      - GetByID\n\n# The end.\n",
			steps: []addStep{
				step("add service Billing --deps UserRepo --methods Charge",
					addsService(config.Service{Name: "Billing", Deps: []string{"UserRepo"}, Methods: []string{"Charge"}})),
			},
			want: func(cfg *config.Config) {
				cfg.Services = config.Services{{Name: "Billing", Methods: []string{"Charge"}, Deps: []string{"UserRepo"}}}
			},
		},
		{
			name: "no adapters key",
			yaml: minimalApp + "\nstorages:\n  - mysql\n\n# The end.\n",
			steps: []addStep{
				step("add adapter UserRepo --storage mysql --methods GetByID",
					addsAdapter(config.Adapter{Name: "UserRepo", Storage: "mysql", Methods: []string{"GetByID"}})),
			},
			want: func(cfg *config.Config) {
				cfg.Adapters = config.Adapters{{Name: "UserRepo", Storage: "mysql", Methods: []string{"GetByID"}}}
			},
		},
		{
			name: "no storages and no adapters key",
			yaml: minimalApp,
			steps: []addStep{
				step("add adapter PaymentsRepo --storage pgsqlx --methods Save",
					addsAdapter(config.Adapter{Name: "PaymentsRepo", Storage: "pgsqlx", Methods: []string{"Save"}})),
			},
			want: func(cfg *config.Config) {
				cfg.Storages = config.Storages{"pgsqlx"}
				cfg.Adapters = config.Adapters{{Name: "PaymentsRepo", Storage: "pgsqlx", Methods: []string{"Save"}}}
			},
		},
		{
			name: "no use_case.methods",
			yaml: minimalApp + "\nuse_case:\n  deps:\n    - Users\n\nservices:\n  - name: Users\n",
			steps: []addStep{
				step("add usecase SignOut", addsUseCase("SignOut")),
			},
			want: func(cfg *config.Config) {
				cfg.UseCase.Methods = []string{"SignOut"}
			},
		},
		{
			name: "no use_case key",
			yaml: minimalApp + "\n# The end.\n",
			steps: []addStep{
				step("add usecase SignOut", addsUseCase("SignOut")),
			},
			want: func(cfg *config.Config) {
				cfg.UseCase.Methods = []string{"SignOut"}
			},
		},
		{
			name: "a methods key without items",
			yaml: minimalApp + "\nuse_case:\n  methods:\n\nstorages:\n  - mysql\n",
			steps: []addStep{
				step("add usecase SignOut", addsUseCase("SignOut")),
			},
			want: func(cfg *config.Config) {
				cfg.UseCase.Methods = []string{"SignOut"}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { runGrowCase(t, c) })
	}
}

// unindentedYaml is testdata/add/goro.yaml in the other common layout: list items at the same
// indentation as their key.
const unindentedYaml = `# Lists here start at the column of their key.
app:
  name: demo
  module: demo
  work_dir: .

storages:
- mysql
# - pgsqlx

use_case:
  deps:
  - Users
  methods:
  - SignIn
  - Pong

services:
- name: Users
  methods:
  - Create
  deps:
  - UserRepo

adapters:
- name: UserRepo
  storage: mysql
  methods:
  - GetByID
`

func TestAddFollowsTheFileLayout(t *testing.T) {
	cases := []growCase{
		{
			// Items inserted at a fixed indentation would continue the scalar above them.
			name:  "list items at the indentation of their key",
			yaml:  unindentedYaml,
			steps: addSteps,
			want: func(cfg *config.Config) {
				*cfg = afterAddSteps(*cfg)
			},
		},
		{
			// Appending to the last line without a newline gives "- Pong    - SignOut", one
			// scalar.
			name: "a use case method list on the last line, without a final newline",
			yaml: minimalApp + "\nstorages:\n  - mysql\n\nuse_case:\n  methods:\n    - SignIn\n    - Pong",
			steps: []addStep{
				step("add usecase SignOut", addsUseCase("SignOut")),
			},
			want: func(cfg *config.Config) {
				cfg.UseCase.Methods = append(cfg.UseCase.Methods, "SignOut")
			},
		},
		{
			name: "an adapter on the last line, without a final newline",
			yaml: minimalApp + "\nstorages:\n  - mysql\n\nadapters:\n  - name: UserRepo\n    storage: mysql\n    methods:\n      - GetByID",
			steps: []addStep{
				step("add method UserRepo.Save", addsMethod("UserRepo.Save")),
				step("add adapter AuditRepo --storage mysql --methods Write",
					addsAdapter(config.Adapter{Name: "AuditRepo", Storage: "mysql", Methods: []string{"Write"}})),
			},
			want: func(cfg *config.Config) {
				cfg.Adapters[0].Methods = append(cfg.Adapters[0].Methods, "Save")
				cfg.Adapters = append(cfg.Adapters, config.Adapter{Name: "AuditRepo", Storage: "mysql", Methods: []string{"Write"}})
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { runGrowCase(t, c) })
	}
}
