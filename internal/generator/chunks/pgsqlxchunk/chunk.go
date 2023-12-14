package pgsqlxchunk

import (
	_ "embed"

	"github.com/hanagantig/goro/internal/config"
)

//go:embed build.tpl
var buildTmpl string

const name = "pgsqlx"
const initName = "pgSqlxConn"

func NewPostgresChunk() config.Chunk {
	return config.Chunk{
		Name:              name,
		Scope:             "storage.database.pgsqlx",
		ArgName:           initName,
		ReturnType:        "*sqlx.DB",
		DefinitionImports: "\"github.com/jmoiron/sqlx\"",
		BuildImports:      "_ \"github.com/lib/pq\"\n\"github.com/jmoiron/sqlx\"\n\"strings\"",
		InitFunc:          "newPgSqlxConnect",
		Build:             buildTmpl,
		Configs:           "pgSqlx configs",
		InitConfig:        "cfg.MainDB",
		InitHasErr:        true,
	}
}
