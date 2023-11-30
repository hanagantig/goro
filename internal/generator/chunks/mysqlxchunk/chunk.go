package mysqlxchunk

import (
	_ "embed"

	"github.com/hanagantig/goro/internal/config"
)

//go:embed build.tpl
var buildTmpl string

const name = "mysqlx"
const initName = "mysqlxConn"

func NewMySQLxChunk() config.Chunk {
	return config.Chunk{
		Name:              name,
		Scope:             "storage.database.mysqlx",
		ArgName:           initName,
		ReturnType:        "*sqlx.DB",
		DefinitionImports: "\"github.com/jmoiron/sqlx\"",
		BuildImports:      "_ \"github.com/go-sql-driver/mysql\"\n\"github.com/jmoiron/sqlx\"\n\"net/url\"\n\"strconv\"\n\"strings\"",
		InitFunc:          "newMySQLxConnect",
		Build:             buildTmpl,
		Configs:           "mysqlx configs",
	}
}
