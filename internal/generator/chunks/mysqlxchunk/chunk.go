package mysqlxchunk

import (
	_ "embed"
	"github.com/hanagantig/goro/internal/config"
)

//go:embed build.tpl
var buildTmpl string

const name = "mysqlx"
const initName = "mysqlxConn"
const initType = "github.com/jmoiron/sqlx"
const initHasErr = true

func NewMySQLxChunk() config.Chunk {
	return config.Chunk{
		Name:              name,
		Scope:             "storage",
		ArgName:           initName,
		ReturnType:        "*sqlx.DB",
		DefinitionImports: "\"github.com/jmoiron/sqlx\"",
		InitFunc:          "newMySQLxConnect",
		Build:             buildTmpl,
		Configs:           "mysqlx configs",
	}
}
