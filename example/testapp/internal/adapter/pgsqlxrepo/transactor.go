package pgsqlxrepo

import (
	"context"

	"errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type contextKey string

const txKey contextKey = "sql_tx"
const txIDKey contextKey = "tx_id"

type Transactor struct {
	conn  *sqlx.DB
	wraps map[context.Context][]func(ctx context.Context) error
}

func NewTransactor(c *sqlx.DB) Transactor {
	return Transactor{conn: c}
}

func (t *Transactor) NewTxContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, txIDKey, uuid.NewString())
}

func (t *Transactor) hasTxID(ctx context.Context) bool {
	txID := ctx.Value(txIDKey)
	return txID != nil && txID != ""
}

func (t *Transactor) InTransaction(ctx context.Context, txFunc func(ctx context.Context) error) error {
	if !t.hasTxID(ctx) {
		return errors.New("not transaction context. Please create it with NewTxContext")
	}

	if _, ok := t.wraps[ctx]; !ok {
		t.wraps[ctx] = make([]func(ctx context.Context) error, 0, 0)
	}

	t.wraps[ctx] = append(t.wraps[ctx], txFunc)

	return nil
}

func (t *Transactor) GetConn(ctx context.Context) *sqlx.DB {
	conn, ok := ctx.Value(txKey).(*sqlx.DB)
	if !ok {
		return t.conn
	}

	return conn
}

func (t *Transactor) RunTransaction(ctx context.Context) error {
	defer t.reset(ctx)

	tx, err := t.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, txKey, tx)
	for _, wrap := range t.wraps[ctx] {
		err = wrap(txCtx)
		if err != nil {
			err = tx.Rollback()
			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (t *Transactor) reset(ctx context.Context) {
	delete(t.wraps, ctx)
}
