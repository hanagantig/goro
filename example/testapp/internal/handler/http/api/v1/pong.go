package v1

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func (h Handler) Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	res, err := h.uc.Pong(ctx)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		h.logger.Error("failed to encode json", slog.String("err", err.Error()))
	}
}
