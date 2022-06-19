package commands

import (
	"goro/internal/config"
	"goro/internal/generator"
	"goro/internal/generator/chains"
	"goro/internal/pkg/log"
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

	g := generator.NewGenerator()

	g.AddChain(chains.NewFitFileNameChain(cfg))
	g.AddChain(chains.NewFitFileExtensionChain(cfg))
	g.AddChain(chains.NewGenerateCodeChain(cfg))
	g.AddChain(chains.NewModInitChain(cfg))
	g.AddChain(chains.NewModTidyChain(cfg.App.WorkDir))

	err = g.Generate()
	_ = err
}
