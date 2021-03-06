package mysqlchunk

import (
	_ "embed"
	"github.com/hanagantig/goro/internal/config"
)

//go:embed build.tpl
var buildTmpl string

const name = "mysql"
const initName = "mysqlConn"
const initType = "*sql.DB"
const initHasErr = true

func NewMySQLChunk() config.Chunk {
	return config.Chunk{
		Name:              name,
		Scope:             "storage",
		ArgName:           "mysqlConnect",
		ReturnType:        "*sql.DB",
		DefinitionImports: "\"database/sql\"",
		BuildImports:      "\"database/sql\"",
		InitFunc:          "newMySQLConnect",
		Build:             buildTmpl,
		Configs:           "mysql configs",
	}
}
