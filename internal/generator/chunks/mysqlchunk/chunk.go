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

type MysqlChunck struct{}

func (m *MysqlChunck) GetName() string {
	return name
}

func (m *MysqlChunck) GetDefinitionImports() string {
	return "imports"
}

func (m *MysqlChunck) GetInit() string {
	return initName
}

func (m *MysqlChunck) GetBuild() string {
	return buildTmpl
}

func (m *MysqlChunck) GetConfig() string {
	return ""
}

func NewMySQLChunk() *MysqlChunck {
	return &MysqlChunck{}
}
