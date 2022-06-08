package commands

import (
	"goro/internal/entity"
	"goro/internal/generator"
	"goro/internal/generator/chains"
)

func InitApp() {
	appData := entity.AppData{}
	err := appData.AskAndSetName()
	if err != nil {
		return
	}

	err = appData.AskAndSetWorkDir()
	if err != nil {
		return
	}

	g := generator.NewGenerator()

	g.AddChain(chains.NewBasementChain(appData))
	g.AddChain(chains.NewFitFileNameChain(appData))
	g.AddChain(chains.NewFitFileExtensionChain(appData))
	g.AddChain(chains.NewGenerateCodeChain(appData))
	g.AddChain(chains.NewModInitChain(appData))
	g.AddChain(chains.NewModTidyChain(appData.WorkDir))

	err = g.Generate()
	_ = err
}
