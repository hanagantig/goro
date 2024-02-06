package generator

import (
	"github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/internal/generator/chunks/httpchunk"
	"github.com/hanagantig/goro/internal/generator/chunks/mysqlchunk"
	"github.com/hanagantig/goro/internal/generator/chunks/mysqlxchunk"
	"github.com/hanagantig/goro/internal/generator/chunks/pgsqlxchunk"
	"github.com/hanagantig/goro/internal/generator/chunks/slogchunk"
	"github.com/hanagantig/goro/internal/generator/chunks/zapchunk"
)

var supportedChunks = map[string]config.Chunk{
	"mysql":  mysqlchunk.NewMySQLChunk(),
	"mysqlx": mysqlxchunk.NewMySQLxChunk(),
	"pgsqlx": pgsqlxchunk.NewPostgresChunk(),
	"http":   httpchunk.NewHttpChunk(),
	"zap":    zapchunk.NewZapLoggerChunk(),
	"slog":   slogchunk.NewSlogLoggerChunk(),
}
