package generator

import (
	"github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/internal/generator/chunks/httpchunk"
	"github.com/hanagantig/goro/internal/generator/chunks/mysqlchunk"
	"github.com/hanagantig/goro/internal/generator/chunks/mysqlxchunk"
	"github.com/hanagantig/goro/internal/generator/chunks/pgsqlxchunk"
)

var supportedChunks = map[string]config.Chunk{
	"mysql":  mysqlchunk.NewMySQLChunk(),
	"mysqlx": mysqlxchunk.NewMySQLxChunk(),
	"pgsqlx": pgsqlxchunk.NewPostgresChunk(),
	"http":   httpchunk.NewHttpChunk(),
}
