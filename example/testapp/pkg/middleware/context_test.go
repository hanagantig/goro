package middleware

import (
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextMiddleware_RequestIDHeaderNotFound(t *testing.T) {
	handler := panicHandler{}
	middleware := AddContextMiddleware(zap.NewNop())

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	middleware(handler).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `request id header not found`)
}
