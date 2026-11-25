package commands_test

import (
	"io"
	"testing"

	"github.com/hanagantig/goro/internal/commands"
	"github.com/hanagantig/goro/internal/config"
)

// TestUpdateRejectsBadConfig: update refuses a goro.yaml that breaks a rule, names the value at
// fault and writes nothing. The rules:
//
//   - service and adapter names follow the method name rule;
//   - names are unique across services and adapters, and so are the files and packages they map
//     to;
//   - a method's file can't be a dep file, a file goro owns, or a file Go builds only for tests
//     or for one OS or architecture;
//   - a transactional adapter's methods can't clash with the Transactor its Repository embeds;
//   - every adapter has a storage, and every storage is one goro supports;
//   - a service doesn't depend on itself.
//
// The project is baseConfig: the Users service depends on UserRepo, a mysql adapter, so
// transactional.
func TestUpdateRejectsBadConfig(t *testing.T) {
	cases := []struct {
		name  string
		edit  func(cfg *config.Config)
		names []string
		// mentions lists other text the error holds.
		mentions []string
	}{
		// Unique names, files and packages.
		{
			name: "a service and an adapter with the same name",
			edit: func(cfg *config.Config) {
				cfg.Adapters = append(cfg.Adapters, config.Adapter{Name: "Users", Storage: "mysql"})
			},
			names: []string{"Users"},
		},
		{
			name: "two services with the same name",
			edit: func(cfg *config.Config) {
				cfg.Services = append(cfg.Services, config.Service{Name: "Billing", Methods: []string{"Refund"}})
			},
			names: []string{"Billing"},
		},
		{
			name: "two adapters with the same name",
			edit: func(cfg *config.Config) {
				cfg.Adapters = append(cfg.Adapters, config.Adapter{Name: "UserRepo", Storage: "mysql"})
			},
			names: []string{"UserRepo"},
		},
		{
			// Both get internal/adapter/mysqlrepo/userrepo.
			name: "two adapters with the same package",
			edit: func(cfg *config.Config) {
				cfg.Adapters = append(cfg.Adapters, config.Adapter{Name: "Userrepo", Storage: "mysql"})
			},
			names: []string{"Userrepo"},
		},
		{
			// container.go imports both packages as userrepo.
			name: "a service and an adapter with the same package name",
			edit: func(cfg *config.Config) {
				cfg.Services = append(cfg.Services, config.Service{Name: "Userrepo"})
			},
			names: []string{"Userrepo"},
		},
		{
			// Both get internal/usecase/sign_in.go.
			name: "two use case methods with the same file",
			edit: func(cfg *config.Config) {
				cfg.UseCase.Methods = append(cfg.UseCase.Methods, "SignIN")
			},
			names: []string{"SignIN"},
		},

		// Bad names.
		{
			name: "a service name that doesn't start with a capital letter",
			edit: func(cfg *config.Config) {
				cfg.Services = append(cfg.Services, config.Service{Name: "notifier", Methods: []string{"Send"}})
			},
			names: []string{"notifier"},
		},
		{
			name: "an adapter name with a character other than a letter",
			edit: func(cfg *config.Config) {
				cfg.Adapters = append(cfg.Adapters, config.Adapter{Name: "AuditRepo2", Storage: "mysql"})
			},
			names: []string{"AuditRepo2"},
		},

		// Method files.
		{
			// DepUsers would get internal/usecase/dep_users.go, the dep file of Users.
			name: "a use case method whose file is a dep file",
			edit: func(cfg *config.Config) {
				cfg.UseCase.Methods = append(cfg.UseCase.Methods, "DepUsers")
			},
			names: []string{"DepUsers"},
		},
		{
			// DepUserRepo would get internal/service/users/dep_user_repo.go, the dep file of
			// UserRepo.
			name:  "a service method whose file is a dep file",
			edit:  func(cfg *config.Config) { addServiceMethod(cfg, "Users", "DepUserRepo") },
			names: []string{"DepUserRepo"},
		},
		{
			// Usecase would get internal/usecase/usecase.go, which goro owns.
			name: "a use case method whose file is goro's",
			edit: func(cfg *config.Config) {
				cfg.UseCase.Methods = append(cfg.UseCase.Methods, "Usecase")
			},
			names: []string{"Usecase"},
		},
		{
			// Service would get internal/service/users/service.go, which goro owns.
			name:  "a service method whose file is goro's",
			edit:  func(cfg *config.Config) { addServiceMethod(cfg, "Users", "Service") },
			names: []string{"Service"},
		},
		{
			// Repository would get internal/adapter/mysqlrepo/userrepo/repository.go, which goro
			// owns.
			name:  "an adapter method whose file is goro's",
			edit:  func(cfg *config.Config) { addAdapterMethod(cfg, "UserRepo", "Repository") },
			names: []string{"Repository"},
		},
		{
			// SignTest would get sign_test.go, which go build leaves out.
			name: "a use case method whose file is a test file",
			edit: func(cfg *config.Config) {
				cfg.UseCase.Methods = append(cfg.UseCase.Methods, "SignTest")
			},
			names: []string{"SignTest"},
		},
		{
			// ReadLinux would get read_linux.go, which builds only on Linux.
			name:  "a service method whose file builds only on one OS",
			edit:  func(cfg *config.Config) { addServiceMethod(cfg, "Users", "ReadLinux") },
			names: []string{"ReadLinux"},
		},
		{
			// FetchArm would get fetch_arm.go, which builds only on arm.
			name:  "an adapter method whose file builds only on one architecture",
			edit:  func(cfg *config.Config) { addAdapterMethod(cfg, "UserRepo", "FetchArm") },
			names: []string{"FetchArm"},
		},

		// The Transactor a transactional adapter's Repository embeds: the field and the methods
		// it promotes.
		{
			name:  "an adapter method named like its transactor",
			edit:  func(cfg *config.Config) { addAdapterMethod(cfg, "UserRepo", "Transactor") },
			names: []string{"Transactor"},
		},
		{
			name:  "an adapter method that shadows Transactor.NewTxContext",
			edit:  func(cfg *config.Config) { addAdapterMethod(cfg, "UserRepo", "NewTxContext") },
			names: []string{"NewTxContext"},
		},
		{
			name:  "an adapter method that shadows Transactor.InTransaction",
			edit:  func(cfg *config.Config) { addAdapterMethod(cfg, "UserRepo", "InTransaction") },
			names: []string{"InTransaction"},
		},
		{
			name:  "an adapter method that shadows Transactor.GetConn",
			edit:  func(cfg *config.Config) { addAdapterMethod(cfg, "UserRepo", "GetConn") },
			names: []string{"GetConn"},
		},
		{
			name:  "an adapter method that shadows Transactor.RunTransaction",
			edit:  func(cfg *config.Config) { addAdapterMethod(cfg, "UserRepo", "RunTransaction") },
			names: []string{"RunTransaction"},
		},

		// Storages and dependencies.
		{
			name: "an adapter without a storage",
			edit: func(cfg *config.Config) {
				cfg.Adapters = append(cfg.Adapters, config.Adapter{Name: "AuditRepo"})
			},
			names:    []string{"AuditRepo"},
			mentions: []string{"storage"},
		},
		{
			// update already fails here today, while rendering and before writing, but its
			// error doesn't quote the storage.
			name: "a storage goro doesn't support",
			edit: func(cfg *config.Config) {
				cfg.Storages = append(cfg.Storages, "mongo")
			},
			names: []string{"mongo"},
		},
		{
			name:  "a service that depends on itself",
			edit:  func(cfg *config.Config) { addServiceDep(cfg, "Users", "Users") },
			names: []string{"Users"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			dir := t.TempDir()
			initProject(t, newConfig(t, dir))
			ageFiles(t, dir)
			before := snapshot(t, dir)

			cfg := newConfig(t, dir)
			c.edit(&cfg)
			_, err := commands.Update(cfg, commands.Options{SkipGoMod: true, Out: io.Discard})
			assertError(t, err, errorCheck{names: c.names, mentions: c.mentions})
			assertNoChanges(t, before, snapshot(t, dir))
		})
	}
}

func addServiceMethod(cfg *config.Config, service, method string) {
	for i := range cfg.Services {
		if cfg.Services[i].Name == service {
			cfg.Services[i].Methods = append(cfg.Services[i].Methods, method)
		}
	}
}

func addServiceDep(cfg *config.Config, service, dep string) {
	for i := range cfg.Services {
		if cfg.Services[i].Name == service {
			cfg.Services[i].Deps = append(cfg.Services[i].Deps, dep)
		}
	}
}

func addAdapterMethod(cfg *config.Config, adapter, method string) {
	for i := range cfg.Adapters {
		if cfg.Adapters[i].Name == adapter {
			cfg.Adapters[i].Methods = append(cfg.Adapters[i].Methods, method)
		}
	}
}
