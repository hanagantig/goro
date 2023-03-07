package clientrepo

import (
	"github.com/jmoiron/sqlx"
	"testapp/internal/adapter/mysqlxrepo"
)

type Repository struct {
	mysqlxrepo.Transactor
}

func NewRepository(conn *sqlx.DB) *Repository {
	return &Repository{
		mysqlxrepo.NewTransactor(conn),
	}
}
