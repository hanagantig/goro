package generator

import (
	"github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/internal/generator/chunks/mysqlchunk"
	"github.com/hanagantig/goro/internal/generator/chunks/mysqlxchunk"
)

var supportedChunks = map[string]config.Chunk{
	"mysql":  mysqlchunk.NewMySQLChunk(),
	"mysqlx": mysqlxchunk.NewMySQLxChunk(),
}
