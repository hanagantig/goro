package slogchunk

import (
	_ "embed"

	"github.com/hanagantig/goro/internal/config"
)

//go:embed build.tpl
var buildTmpl string

//go:embed pkginterface.tpl
var pkgInterface string

const name = "slog"
const initName = "l"

func NewSlogLoggerChunk() config.Chunk {
	return config.Chunk{
		Name:              name,
		Scope:             "logger.slog",
		ArgName:           initName,
		ReturnType:        "*slog.Logger",
		DefinitionImports: "\"context\"",
		BuildImports:      "\"log/slog\"\n\"os\"\n\"context\"",
		InitFunc:          "NewLogger",
		Build:             buildTmpl,
		Configs:           "logger configs",
		InitConfig:        "",
		PkgInterface:      pkgInterface,
		InitHasErr:        false,
	}
}
