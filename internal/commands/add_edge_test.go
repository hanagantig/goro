package commands_test

import (
	"strings"
	"testing"

	"github.com/hanagantig/goro/internal/config"
)

// TestAddEdgeLayouts: goro add on goro.yaml layouts the review found (see add_layout_test.go for
// what each case checks).
func TestAddEdgeLayouts(t *testing.T) {
	base := readTestdata(t, "add/goro.yaml")
	const methods = "  methods:\n    - SignIn\n"

	cases := []growCase{
		{
			name:  "use_case.methods with an anchor",
			yaml:  strings.Replace(base, methods, "  methods: &ucm\n    - SignIn\n", 1),
			steps: []addStep{step("add usecase SignOut", addsUseCase("SignOut"))},
			want: func(cfg *config.Config) {
				cfg.UseCase.Methods = append(cfg.UseCase.Methods, "SignOut")
			},
		},
		{
			name:  "use_case.methods with explicit tags",
			yaml:  strings.Replace(base, methods, "  methods: !!seq\n    - !!str SignIn\n", 1),
			steps: []addStep{step("add usecase SignOut", addsUseCase("SignOut"))},
			want: func(cfg *config.Config) {
				cfg.UseCase.Methods = append(cfg.UseCase.Methods, "SignOut")
			},
		},
		{
			// Both inserts go after the last line: the adapter has to stay in adapters.
			name: "no storages key and an empty adapters key at the end",
			yaml: minimalApp + "\nadapters:\n",
			steps: []addStep{step("add adapter AuditAPI --storage http --methods Send",
				addsAdapter(config.Adapter{Name: "AuditAPI", Storage: "http", Methods: []string{"Send"}}))},
			want: func(cfg *config.Config) {
				cfg.Storages = config.Storages{"http"}
				cfg.Adapters = config.Adapters{{Name: "AuditAPI", Storage: "http", Methods: []string{"Send"}}}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { runGrowCase(t, c) })
	}
}
