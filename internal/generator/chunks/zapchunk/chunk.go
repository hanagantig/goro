package zapchunk

import (
	_ "embed"

	"github.com/hanagantig/goro/internal/config"
)

//go:embed build.tpl
var buildTmpl string

//go:embed pkginteface.tpl
var pkgInterface string

const name = "zap"
const initName = "l"

func NewZapLoggerChunk() config.Chunk {
	return config.Chunk{
		Name:              name,
		Scope:             "logger.zap",
		ArgName:           initName,
		ReturnType:        "*zap.Logger",
		DefinitionImports: "\"go.uber.org/zap/zapcore\"",
		BuildImports:      "\"go.uber.org/zap/zapcore\"\n\"go.uber.org/zap\"",
		InitFunc:          "NewLogger",
		Build:             buildTmpl,
		Configs:           "logger configs",
		InitConfig:        "",
		PkgInterface:      pkgInterface,
		InitHasErr:        true,
	}
}
