package generator

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"

	entity "github.com/hanagantig/goro/internal/config"
)

var FuncMap = template.FuncMap{
	"renderImports":                  RenderImports,
	"renderDefinition":               RenderDefinitions,
	"renderInitializationsWithError": RenderInitializationsWithError,
	"renderStructPopulation":         RenderStructPopulation,
	"renderArgs":                     RenderArgs,
	"renderBuild":                    RenderBuild,
	"toCamelCase":                    strcase.ToCamel,
	"toPrivateName":                  ToPrivateName,
	"toPublicName":                   ToPublicName,
	"contains":                       strings.Contains,
}

func RenderImports(scope, stage string, cfg entity.Config) string {
	chunks := cfg.GetChunksByScope(scope)
	res := strings.Builder{}
	for _, ch := range chunks {
		switch stage {
		case "build":
			fmt.Fprintf(&res, "%v\n", ch.BuildImports)
		case "definition":
			fmt.Fprintf(&res, "%v\n", ch.DefinitionImports)
		default:
			fmt.Fprintf(&res, "%v\n", ch.DefinitionImports)
		}
	}

	return res.String()
}

func RenderDefinitions(scope string, cfg entity.Config) string {
	chunks := cfg.GetChunksByScope(scope)
	res := strings.Builder{}
	for _, ch := range chunks {
		fmt.Fprintf(&res, "%v %v\n", ch.Name, ch.ReturnType)
	}

	return res.String()
}

func RenderInitializationsWithError(scope, prefix string, cfg entity.Config) string {
	chunks := cfg.GetChunksByScope(scope)
	res := strings.Builder{}
	for _, ch := range chunks {
		fmt.Fprintf(&res, "%v,%v := %v.%v(cfg.MainDB)\n", ch.ArgName, "err", prefix, ch.InitFunc)
		fmt.Fprintf(&res, "if err != nil {\n return nil, err\n}\n")
		fmt.Fprintf(&res, "%v.%v = %v\n", prefix, ch.Name, ch.ArgName)
	}

	return res.String()
}

func RenderDependency(scope, prefix string, cfg entity.Config) string {
	return "// render dependencies code"
}

func RenderBuild(scope string, cfg entity.Config) string {
	chunks := cfg.GetChunksByScope(scope)
	res := strings.Builder{}
	for _, ch := range chunks {
		fmt.Fprintf(&res, "%v\n\n", ch.Build)
	}

	return res.String()
}

func RenderArgs(scope string, cfg entity.Config) string {
	chunks := cfg.GetChunksByScope(scope)
	res := strings.Builder{}
	for _, ch := range chunks {
		fmt.Fprintf(&res, "%v %v,", ch.ArgName, ch.ReturnType)
	}

	return res.String()
}

func RenderStructPopulation(scope string, cfg entity.Config) string {
	chunks := cfg.GetChunksByScope(scope)
	res := strings.Builder{}
	for _, ch := range chunks {
		fmt.Fprintf(&res, "%v: %v,\n", ch.Name, ch.ArgName)
	}

	return res.String()
}

func ToPrivateName(name string) string {
	return strings.ToLower(string(name[0])) + name[1:]
}

func ToPublicName(name string) string {
	return strings.ToUpper(string(name[0])) + name[1:]
}
