package generator

import (
	"goro/internal/config"
	"goro/internal/generator/chunks/mysqlchunk"
)

var supportedChunks = map[string]config.Chunk{
	"mysql": mysqlchunk.NewMySQLChunk(),
}
