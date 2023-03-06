package v1

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"testapp/pkg/logger"
)

type ExampleUseCase interface {
	Pong(ctx context.Context) (string, error)
}

type Handler struct {
	uc     ExampleUseCase
	logger logger.Logger
}

func NewHandler(uc ExampleUseCase, logs logger.Logger) *Handler {
	return &Handler{uc: uc, logger: logs}
}

func (h Handler) Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	res, err := h.uc.Pong(ctx)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		h.logger.Error("failed to encode json", zap.Error(err))
	}
}
