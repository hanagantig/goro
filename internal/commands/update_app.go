package commands

import (
	"github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/internal/generator"
	"github.com/hanagantig/goro/internal/generator/chains"
)

// Update brings the project in cfg.App.WorkDir in line with cfg.
func Update(cfg config.Config, opts Options) (Report, error) {
	err := cfg.Validate()
	if err != nil {
		return Report{}, err
	}

	g := generator.NewGenerator(cfg)

	g.AddChain(chains.NewFitFileNameChain())
	g.AddChain(chains.NewGenerateAdapterChain())
	g.AddChain(chains.NewGenerateServicesChain())
	g.AddChain(chains.NewGenerateUseCaseChain())
	g.AddChain(chains.NewGenerateCodeChain())
	g.AddChain(chains.NewFitFileExtensionChain())

	g.AddChain(chains.NewSyncUseCaseChain())
	g.AddChain(chains.NewSyncServicesChain())
	g.AddChain(chains.NewSyncAdaptersChain())
	g.AddChain(chains.NewUpdateFilesChain())
	if !opts.SkipGoMod {
		g.AddChain(chains.NewModTidyChain())
	}

	return Report{}, g.Generate()
}
