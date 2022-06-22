package mysqlchunk

import (
	_ "embed"
)

//go:embed build.tpl
var buildTmpl string

const name = "mysql"
const initName = "mysqlConn"
const initType = "*sql.DB"
const initHasErr = true

type MysqlChunk struct{}

func (m *MysqlChunk) GetName() string {
	return name
}

func (m *MysqlChunk) GetDefinitionImports() string {
	return "imports"
}

func (m *MysqlChunk) GetInit() string {
	return initName
}

func (m *MysqlChunk) GetBuild() string {
	return buildTmpl
}

func (m *MysqlChunk) GetConfig() string {
	return ""
}

func NewMySQLChunk() *MysqlChunk {
	return &MysqlChunk{}
}
