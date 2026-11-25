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
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { runGrowCase(t, c) })
	}
}
