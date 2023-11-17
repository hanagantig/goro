package httpchunk

import (
	_ "embed"
	"github.com/hanagantig/goro/internal/config"
)

//go:embed build.tpl
var buildTmpl string

const name = "http"
const initName = "httpClient"

func NewHttpChunk() config.Chunk {
	return config.Chunk{
		Name:              name,
		Scope:             "storage",
		ArgName:           initName,
		ReturnType:        "*http.Client",
		DefinitionImports: "\"net/http\"",
		BuildImports:      "client \"net/http\"",
		InitFunc:          "newHttpClient",
		Build:             buildTmpl,
		Configs:           "http configs",
	}
}
