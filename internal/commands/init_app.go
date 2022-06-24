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

	g := generator.NewGenerator(cfg)

	g.AddChain(chains.NewFitFileNameChain())
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
