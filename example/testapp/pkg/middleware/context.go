package middleware

import (
	"context"
	"net/http"
	"testapp/pkg/logger"
)

func AddContextMiddleware(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			//if requestID == "" {
			//	w.WriteHeader(http.StatusBadRequest)
			//	_, err := w.Write([]byte(`request id header not found`))
			//	if err != nil {
			//		log.Error("failed to write response body")
			//	}
			//	return
			//}
			_ = requestID
			ctx := context.Background()
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
