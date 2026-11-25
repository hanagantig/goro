package commands_test

import (
	"io"
	"strings"
	"testing"

	"github.com/hanagantig/goro/internal/commands"
	"github.com/hanagantig/goro/internal/config"
)

// TestUpdateRejectsClashingNames: names are unique across services and adapters, and a method's
// file can't be a dep file. update refuses such a goro.yaml, names the value at fault and writes
// nothing.
func TestUpdateRejectsClashingNames(t *testing.T) {
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
			name: "a service method whose file is a dep file",
			edit: func(cfg *config.Config) {
				for i := range cfg.Services {
					if cfg.Services[i].Name == "Users" {
						cfg.Services[i].Methods = append(cfg.Services[i].Methods, "DepUserRepo")
					}
				}
			},
			inErr: "DepUserRepo",
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
