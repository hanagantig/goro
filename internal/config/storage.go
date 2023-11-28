package config

var storagePackages = map[Storage]string{
	"mysql":  "\"database/sql\"",
	"mysqlx": "\"github.com/jmoiron/sqlx\"",
	"pgsqlx": "\"github.com/jmoiron/sqlx\"",
}

var connectionsType = map[Storage]string{
	"mysql":  "*sql.DB",
	"mysqlx": "*sqlx.DB",
	"pgsqlx": "*sqlx.DB",
}

var connectionName = map[Storage]string{
	"pgsqlx": "conn",
	"mysql":  "conn",
	"mysqlx": "conn",
}

var transactionalStorages = map[Storage]bool{
	"pgsqlx": true,
	"mysql":  true,
	"mysqlx": true,
}

type Storage string
type Storages []Storage

func (s Storages) GetMap() map[Storage]struct{} {
	res := make(map[Storage]struct{}, 0)
	for _, st := range s {
		res[st] = struct{}{}
	}

	return res
}

func (s Storage) String() string {
	return string(s)
}

func (s Storage) GetFolderName() string {
	return string(s) + "repo"
}

func (s Storage) GetConnImportName() string {
	return storagePackages[s]
}

func (s Storage) GetConnectionType() string {
	return connectionsType[s]
}

func (s Storage) GetConnectionName() string {
	return connectionName[s]
}
