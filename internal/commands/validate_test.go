package commands_test

import (
	"io"
	"strings"
	"testing"

	"github.com/hanagantig/goro/internal/commands"
	"github.com/hanagantig/goro/internal/config"
)

// TestUpdateRejectsBadConfig: service and adapter names follow the method name rule and are unique
// across services and adapters, a method's file can't be a dep file or a file goro owns, and
// every storage is one goro supports. update refuses a goro.yaml that breaks a rule, names the
// value at fault and writes nothing.
func TestUpdateRejectsBadConfig(t *testing.T) {
	cases := []struct {
		name  string
		edit  func(cfg *config.Config)
		inErr string
	}{
		{
			name: "a service and an adapter with the same name",
			edit: func(cfg *config.Config) {
				cfg.Adapters = append(cfg.Adapters, config.Adapter{Name: "Users", Storage: "mysql"})
			},
			inErr: "Users",
		},
		{
			name: "two services with the same name",
			edit: func(cfg *config.Config) {
				cfg.Services = append(cfg.Services, config.Service{Name: "Billing", Methods: []string{"Refund"}})
			},
			inErr: "Billing",
		},
		{
			name: "two adapters with the same name",
			edit: func(cfg *config.Config) {
				cfg.Adapters = append(cfg.Adapters, config.Adapter{Name: "UserRepo", Storage: "mysql"})
			},
			inErr: "UserRepo",
		},
		{
			// DepUsers would get internal/usecase/dep_users.go, the dep file of Users.
			name: "a use case method whose file is a dep file",
			edit: func(cfg *config.Config) {
				cfg.UseCase.Methods = append(cfg.UseCase.Methods, "DepUsers")
			},
			inErr: "DepUsers",
		},
		{
			// DepUserRepo would get internal/service/users/dep_user_repo.go, the dep file of
			// UserRepo.
			name:  "a service method whose file is a dep file",
			edit:  func(cfg *config.Config) { addServiceMethod(cfg, "Users", "DepUserRepo") },
			inErr: "DepUserRepo",
		},
		{
			name: "a service name that doesn't start with a capital letter",
			edit: func(cfg *config.Config) {
				cfg.Services = append(cfg.Services, config.Service{Name: "notifier", Methods: []string{"Send"}})
			},
			inErr: "notifier",
		},
		{
			name: "an adapter name with a character other than a letter",
			edit: func(cfg *config.Config) {
				cfg.Adapters = append(cfg.Adapters, config.Adapter{Name: "AuditRepo2", Storage: "mysql"})
			},
			inErr: "AuditRepo2",
		},
		{
			// Usecase would get internal/usecase/usecase.go, which goro owns.
			name: "a use case method whose file is goro's",
			edit: func(cfg *config.Config) {
				cfg.UseCase.Methods = append(cfg.UseCase.Methods, "Usecase")
			},
			inErr: "Usecase",
		},
		{
			// Service would get internal/service/users/service.go, which goro owns.
			name:  "a service method whose file is goro's",
			edit:  func(cfg *config.Config) { addServiceMethod(cfg, "Users", "Service") },
			inErr: "Service",
		},
		{
			// Repository would get internal/adapter/mysqlrepo/userrepo/repository.go, which goro
			// owns.
			name:  "an adapter method whose file is goro's",
			edit:  func(cfg *config.Config) { addAdapterMethod(cfg, "UserRepo", "Repository") },
			inErr: "Repository",
		},
		{
			// Transactor would get transactor.go, the name of goro's transactor files, and clash
			// with the Transactor that a transactional adapter's Repository embeds.
			name:  "an adapter method named like its transactor",
			edit:  func(cfg *config.Config) { addAdapterMethod(cfg, "UserRepo", "Transactor") },
			inErr: "Transactor",
		},
		{
			// Passes today: rendering already fails on a storage without a chunk, before
			// anything is written. It keeps that so when the check moves into validation.
			name: "a storage goro doesn't support",
			edit: func(cfg *config.Config) {
				cfg.Storages = append(cfg.Storages, "mongo")
			},
			inErr: "mongo",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			initProject(t, newConfig(t, dir))
			ageFiles(t, dir)
			before := snapshot(t, dir)

			cfg := newConfig(t, dir)
			c.edit(&cfg)
			_, err := commands.Update(cfg, commands.Options{SkipGoMod: true, Out: io.Discard})
			switch {
			case err == nil:
				t.Errorf("update accepted the goro.yaml, want an error")
			case !strings.Contains(err.Error(), c.inErr):
				t.Errorf("error %q doesn't name %q", err, c.inErr)
			}
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

func addAdapterMethod(cfg *config.Config, adapter, method string) {
	for i := range cfg.Adapters {
		if cfg.Adapters[i].Name == adapter {
			cfg.Adapters[i].Methods = append(cfg.Adapters[i].Methods, method)
		}
	}
}
