package service

import (
	"context"
)

type Transactor interface {
	NewTxContext(ctx context.Context) context.Context
	InTransaction(ctx context.Context, txFunc func(ctx context.Context) error) error
	RunTransaction(ctx context.Context) error
}
