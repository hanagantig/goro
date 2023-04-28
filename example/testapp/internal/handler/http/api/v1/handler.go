package v1

import (
	"context"
	"testapp/pkg/logger"
)

type UseCase interface {
	GetClients(ctx context.Context) (interface{}, error)
	SignIn(ctx context.Context) (interface{}, error)
	SignUp(ctx context.Context) (interface{}, error)
	Pong(ctx context.Context) (interface{}, error)
}

type Handler struct {
	uc     UseCase
	logger logger.Logger
}

func NewHandler(uc UseCase, logs logger.Logger) *Handler {
	return &Handler{uc: uc, logger: logs}
}
