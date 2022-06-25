package generator

import (
	"fmt"
	entity "github.com/hanagantig/goro/internal/config"
	"strings"
	"text/template"
)

var FunkMap = template.FuncMap{
	"renderImports":                  RenderImports,
	"renderDefinition":               RenderDefinitions,
	"renderInitializationsWithError": RenderInitializationsWithError,
	"renderDependency":               RenderDependency,
	"renderStructPopulation":         RenderStructPopulation,
	"renderArgs":                     RenderArgs,
	"renderBuild":                    RenderBuild,
}

func RenderImports(scope string, chunks []entity.Chunk) string {
	res := strings.Builder{}
	for _, ch := range chunks {
		fmt.Fprintf(&res, "%v", ch.DefinitionImports)
	}

	return res.String()
}

func RenderDefinitions(scope string, chunks []entity.Chunk) string {
	res := strings.Builder{}
	for _, ch := range chunks {
		fmt.Fprintf(&res, "%v %v\n", ch.Name, ch.ReturnType)
	}

	return res.String()
}

func RenderInitializationsWithError(scope, prefix string, chunks []entity.Chunk) string {
	res := strings.Builder{}
	for _, ch := range chunks {
		tmpName := "mysqlConn"
		fmt.Fprintf(&res, "%v,%v := %v.newMySQLConnect(cfg.MainDB)\n", tmpName, "err", prefix)
		fmt.Fprintf(&res, "if err != nil {\n return nil, err\n}\n")
		fmt.Fprintf(&res, "%v.%v = %v\n", prefix, ch.Name, tmpName)
	}

	return res.String()
}

func RenderDependency(scope, prefix string, chunks []entity.Chunk) string {
	return "// render dependencies code"
}

func RenderBuild(scope string, chunks []entity.Chunk) string {
	res := strings.Builder{}
	for _, ch := range chunks {
		fmt.Fprintf(&res, "%v\n\n", ch.Build)
	}

	return res.String()
}

func RenderArgs(scope string, chunks []entity.Chunk) string {
	res := strings.Builder{}
	for _, ch := range chunks {
		fmt.Fprintf(&res, "%v %v,", ch.ArgName, ch.ReturnType)
	}

	return res.String()
}

func RenderStructPopulation(scope string, chunks []entity.Chunk) string {
	res := strings.Builder{}
	for _, ch := range chunks {
		fmt.Fprintf(&res, "%v: %v,\n", ch.Name, ch.ArgName)
	}

	return res.String()
}
