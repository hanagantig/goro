package commands

import (
	"fmt"
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

	g := generator.NewGenerator()

	g.AddChain(chains.NewBasementChain(appData))
	g.AddChain(chains.NewFitFileNameChain(appData))
	g.AddChain(chains.NewFitFileExtensionChain(appData))
	g.AddChain(chains.NewGenerateCodeChain(appData))

	err = g.Generate()
	fmt.Println(appData, err)
}
