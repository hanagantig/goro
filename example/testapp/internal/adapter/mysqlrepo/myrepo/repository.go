package myrepo

import (
	"database/sql"
	"testapp/internal/adapter/mysqlrepo"
)

type Repository struct {
	mysqlrepo.Transactor
}

func NewRepository(conn *sql.DB) *Repository {
	return &Repository{
		mysqlrepo.NewTransactor(conn),
	}
}
