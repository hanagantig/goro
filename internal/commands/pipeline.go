package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/internal/generator"
	"github.com/hanagantig/goro/internal/generator/chains"
	"github.com/hanagantig/goro/pkg/afero"
)

// render runs the generator chains for cfg and returns the rendered project in memory.
func render(cfg config.Config) (*afero.Afero, error) {
	g := generator.NewGenerator(cfg)

	g.AddChain(chains.NewFitFileNameChain())
	g.AddChain(chains.NewGenerateAdapterChain())
	g.AddChain(chains.NewGenerateServicesChain())
	g.AddChain(chains.NewGenerateUseCaseChain())
	g.AddChain(chains.NewGenerateDepsChain())
	g.AddChain(chains.NewGenerateCodeChain())
	g.AddChain(chains.NewFitFileExtensionChain())

	return g.Generate()
}

// run renders cfg, plans every file against the work dir, writes the plan and prints a summary.
// Under DryRun it prints the plan instead and writes nothing, go mod files included. command
// names the goro command in the output; goModInit adds `go mod init` before `go mod tidy`. conf,
// when set, is a goro.yaml to write first; it is printed with the plan but isn't part of the
// report.
func run(cfg config.Config, opts Options, command string, goModInit bool, conf *configFile) (Report, error) {
	out := opts.Out
	if out == nil {
		out = os.Stdout
	}

	rendered, err := render(cfg)
	if err != nil {
		return Report{}, err
	}

	files, err := planFiles(rendered, cfg.App.WorkDir)
	if err != nil {
		return Report{}, err
	}
	shown := files
	if conf != nil {
		shown = append([]plannedFile{conf.planned()}, files...)
	}

	if opts.DryRun {
		err = printDryRun(out, shown)
		if err != nil {
			return Report{}, err
		}
		return toReport(files), nil
	}

	if conf != nil {
		err = conf.write()
		if err != nil {
			return Report{}, err
		}
	}
	err = writeFiles(cfg.App.WorkDir, files)
	if err != nil {
		return Report{}, err
	}
	printSummary(out, command, cfg.App.WorkDir, shown)

	if !opts.SkipGoMod {
		err = runGoMod(out, cfg, goModInit)
		if err != nil {
			return Report{}, err
		}
	}

	return toReport(files), nil
}

func runGoMod(out io.Writer, cfg config.Config, withInit bool) error {
	steps := []generator.Chain{chains.NewModTidyChain()}
	if withInit {
		steps = append([]generator.Chain{chains.NewModInitChain()}, steps...)
	}

	for _, step := range steps {
		name := strings.ToLower(step.Name())
		fmt.Fprintf(out, "Running %s\n", name)
		if _, err := step.Apply(nil, cfg); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}

	return nil
}
