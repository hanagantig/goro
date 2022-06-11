package commands

import (
	"goro/internal/entity"
	"goro/internal/generator"
	"goro/internal/generator/chains"
	"goro/internal/pkg/log"
)

func InitApp(configPath string) {
	appData, err := entity.NewAppData(configPath)
	if err != nil {
		log.Fatal(err)
	}

	err = appData.AskAndSetName()
	if err != nil {
		log.Fatal(err)
	}

	err = appData.AskAndSetWorkDir()
	if err != nil {
		log.Fatal(err)
	}

	err = appData.Validate()
	if err != nil {
		log.Fatal(err)
	}

	g := generator.NewGenerator()

	g.AddChain(chains.NewBasementChain(appData))
	g.AddChain(chains.NewFitFileNameChain(appData))
	g.AddChain(chains.NewFitFileExtensionChain(appData))
	g.AddChain(chains.NewGenerateCodeChain(appData))
	g.AddChain(chains.NewModInitChain(appData))
	g.AddChain(chains.NewModTidyChain(appData.App.WorkDir))

	err = g.Generate()
	_ = err
}
