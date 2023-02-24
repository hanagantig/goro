package config

var storagePackages = map[Storage]string{
	"mysql":  "\"database/sql\"",
	"mysqlx": "\"github.com/jmoiron/sqlx\"",
}

var connectionsType = map[Storage]string{
	"mysql":  "*sql.DB",
	"mysqlx": "*sqlx.DB",
}

type Storage string

func (s Storage) String() string {
	return string(s)
}

func (s Storage) GetFolderName() string {
	return string(s) + "repo"
}

func (s Storage) GetPkgNameForImport() string {
	return storagePackages[s]
}

func (s Storage) GetConnectionType() string {
	return connectionsType[s]
}
