// Code generated by goro;

package clientrepo

// This file was generated by the goro tool.

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
