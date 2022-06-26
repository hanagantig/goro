package postgresqlchunk

import (
	_ "embed"
	"github.com/hanagantig/goro/internal/config"
)

//go:embed build.tpl
var buildTmpl string

const name = "postgres"
const initName = "postgresConn"

func NewPostgresChunk() config.Chunk {
	return config.Chunk{
		Name:       name,
		Scope:      "storage",
		ArgName:    initName,
		ReturnType: "*sql.DB",
		//DefinitionImports: "\"github.com/jackc/pgx/v4\"\n\"github.com/jackc/pgx/v4/stdlib\"",
		DefinitionImports: "",
		BuildImports:      "\"github.com/jackc/pgx/v4\"\n\"github.com/jackc/pgx/v4/stdlib\"",
		InitFunc:          "newPostgresConnect",
		Build:             buildTmpl,
		Configs:           "postgres configs",
	}
}
