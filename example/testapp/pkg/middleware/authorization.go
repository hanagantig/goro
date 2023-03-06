package middleware

import (
	"net/http"
)

var (
	apiKey = "test"
)

func AuthorizationMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Auth-API-Key")
			userAgent := r.UserAgent()
			if authHeader != apiKey || userAgent == "" {
				http.Error(w, "Authorization error.", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
