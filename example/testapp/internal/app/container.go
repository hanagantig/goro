package app

import (
	"database/sql"
	"github.com/jmoiron/sqlx"

	"testapp/internal/usecase"

	"testapp/internal/service/myservice"
	"testapp/internal/service/pinpong"

	"testapp/internal/adapter/mysqlrepo/myrepo"
	"testapp/internal/adapter/mysqlxrepo/clientrepo"
)

type Container struct {
	mysql    *sql.DB
	mysqlx   *sqlx.DB
	postgres *sql.DB

	deps map[string]interface{}
}

func NewContainer(mysqlConnect *sql.DB, mysqlxConn *sqlx.DB, postgresConn *sql.DB) *Container {

	return &Container{
		mysql:    mysqlConnect,
		mysqlx:   mysqlxConn,
		postgres: postgresConn,

		deps: make(map[string]interface{}),
	}
}

func (c *Container) GetUseCase() *usecase.UseCase {

	return usecase.NewUseCase(c.getMyService())
}

func (c *Container) getMysql() *sql.DB {
	return c.mysql
}

func (c *Container) getMysqlx() *sqlx.DB {
	return c.mysqlx
}

func (c *Container) getPostgres() *sql.DB {
	return c.postgres
}

func (c *Container) getMyService() *myservice.Service {

	return myservice.NewService(c.getMyRepo())
}

func (c *Container) getPinPong() *pinpong.Service {

	return pinpong.NewService(c.getMyRepo())
}

func (c *Container) getMyRepo() *myrepo.Repository {

	return myrepo.NewRepository(c.getMysql())
}

func (c *Container) getClientRepo() *clientrepo.Repository {

	return clientrepo.NewRepository(c.getMysqlx())
}
