package commands

import (
	"github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/internal/generator"
	"github.com/hanagantig/goro/internal/generator/chains"
)

// Init generates a new project from cfg into cfg.App.WorkDir.
func Init(cfg config.Config, opts Options) error {
	err := cfg.Validate()
	if err != nil {
		return err
	}

	g := generator.NewGenerator(cfg)

	g.AddChain(chains.NewFitFileNameChain())
	g.AddChain(chains.NewGenerateAdapterChain())
	g.AddChain(chains.NewGenerateServicesChain())
	g.AddChain(chains.NewGenerateUseCaseChain())
	g.AddChain(chains.NewGenerateDepsChain())
	g.AddChain(chains.NewGenerateCodeChain())
	g.AddChain(chains.NewFitFileExtensionChain())
	g.AddChain(chains.NewSaveFilesChain())
	if !opts.SkipGoMod {
		g.AddChain(chains.NewModInitChain())
		g.AddChain(chains.NewModTidyChain())
	}

	return g.Generate()
}
