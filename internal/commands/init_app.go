package commands

import (
	"github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/internal/generator"
	"github.com/hanagantig/goro/internal/generator/chains"
	"github.com/hanagantig/goro/internal/pkg/log"
)

func InitApp(configPath string) {
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	err = cfg.AskAndSetName()
	if err != nil {
		log.Fatal(err)
	}

	err = cfg.AskAndSetWorkDir()
	if err != nil {
		log.Fatal(err)
	}

	err = cfg.Validate()
	if err != nil {
		log.Fatal(err)
	}

	g := generator.NewGenerator(cfg)

	g.AddChain(chains.NewFitFileNameChain())
	g.AddChain(chains.NewGenerateAdapterChain())
	g.AddChain(chains.NewGenerateServicesChain())
	g.AddChain(chains.NewGenerateUseCaseChain())
	g.AddChain(chains.NewGenerateCodeChain())
	g.AddChain(chains.NewFitFileExtensionChain())
	g.AddChain(chains.NewSaveFilesChain())
	g.AddChain(chains.NewModInitChain())
	g.AddChain(chains.NewModTidyChain())

	err = g.Generate()

	if err != nil {
		log.Fatal(err)
	}
}
