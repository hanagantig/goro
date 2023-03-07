package app

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"testapp/internal/config"

	"net/url"
	"strconv"
	"strings"
)

func (a *App) newMySQLConnect(cfg config.SQLConfig) (*sql.DB, error) {
	builder := strings.Builder{}
	builder.WriteString(cfg.User)
	builder.WriteByte(':')
	builder.WriteString(cfg.Password)
	builder.WriteString("@tcp(")
	builder.WriteString(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port))
	builder.WriteString(")/")
	builder.WriteString(cfg.DBName)

	builder.WriteString("?timeout=")
	builder.WriteString(cfg.Timeout.String())
	builder.WriteString("&readTimeout=")
	builder.WriteString(cfg.ReadTimeout.String())
	builder.WriteString("&writeTimeout=")
	builder.WriteString(cfg.WriteTimeout.String())
	builder.WriteString("&interpolateParams=")
	builder.WriteString(strconv.FormatBool(cfg.InterpolateParams))
	if cfg.Charset != "" {
		builder.WriteString("&charset=")
		builder.WriteString(cfg.Charset)
	}
	builder.WriteString("&parseTime=")
	builder.WriteString(strconv.FormatBool(cfg.ParseTime))
	if cfg.Collation != "" {
		builder.WriteString("&collation=")
		builder.WriteString(cfg.Collation)
	}

	if cfg.Timezone != "" {
		builder.WriteString("&loc=")
		builder.WriteString(url.QueryEscape(cfg.Timezone))
	}
	dsn := builder.String()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	return db, nil
}

func (a *App) newMySQLxConnect(cfg config.SQLConfig) (*sqlx.DB, error) {
	builder := strings.Builder{}
	builder.WriteString(cfg.User)
	builder.WriteByte(':')
	builder.WriteString(cfg.Password)
	builder.WriteString("@tcp(")
	builder.WriteString(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port))
	builder.WriteString(")/")
	builder.WriteString(cfg.DBName)

	builder.WriteString("?timeout=")
	builder.WriteString(cfg.Timeout.String())
	builder.WriteString("&readTimeout=")
	builder.WriteString(cfg.ReadTimeout.String())
	builder.WriteString("&writeTimeout=")
	builder.WriteString(cfg.WriteTimeout.String())
	builder.WriteString("&interpolateParams=")
	builder.WriteString(strconv.FormatBool(cfg.InterpolateParams))
	if cfg.Charset != "" {
		builder.WriteString("&charset=")
		builder.WriteString(cfg.Charset)
	}
	builder.WriteString("&parseTime=")
	builder.WriteString(strconv.FormatBool(cfg.ParseTime))
	if cfg.Collation != "" {
		builder.WriteString("&collation=")
		builder.WriteString(cfg.Collation)
	}

	if cfg.Timezone != "" {
		builder.WriteString("&loc=")
		builder.WriteString(url.QueryEscape(cfg.Timezone))
	}
	dsn := builder.String()

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	return db, nil
}

func (a *App) newPostgresConnect(cfg config.SQLConfig) (*sql.DB, error) {
	builder := strings.Builder{}
	builder.WriteString(cfg.User)
	builder.WriteByte(':')
	builder.WriteString(cfg.Password)
	builder.WriteString("@tcp(")
	builder.WriteString(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port))
	builder.WriteString(")/")
	builder.WriteString(cfg.DBName)

	builder.WriteString("?timeout=")
	builder.WriteString(cfg.Timeout.String())
	builder.WriteString("&readTimeout=")
	builder.WriteString(cfg.ReadTimeout.String())
	builder.WriteString("&writeTimeout=")
	builder.WriteString(cfg.WriteTimeout.String())
	builder.WriteString("&interpolateParams=")
	builder.WriteString(strconv.FormatBool(cfg.InterpolateParams))
	if cfg.Charset != "" {
		builder.WriteString("&charset=")
		builder.WriteString(cfg.Charset)
	}
	builder.WriteString("&parseTime=")
	builder.WriteString(strconv.FormatBool(cfg.ParseTime))
	if cfg.Collation != "" {
		builder.WriteString("&collation=")
		builder.WriteString(cfg.Collation)
	}

	if cfg.Timezone != "" {
		builder.WriteString("&loc=")
		builder.WriteString(url.QueryEscape(cfg.Timezone))
	}
	dsn := builder.String()

	pgxCfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	db := stdlib.OpenDB(*pgxCfg)

	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	return db, nil
}
