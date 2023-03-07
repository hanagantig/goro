package middleware

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

// responseWriterWrapper is used to wrap http.ResponseWriter.
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

// NewResponseWriterWrapper creates a responseWriterWrapper.
func newResponseWriterWrapper(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader wraps WriteHeader method for http.ResponseWriter.
func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write wraps Write method for http.ResponseWriter.
func (rw *responseWriterWrapper) Write(body []byte) (int, error) {
	rw.body = body
	return rw.ResponseWriter.Write(body)
}

// Hijack wraps Hijack method for http.Hijacker.
// It is used for WebSocket handler.
func (rw *responseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("websocket: response does not implement http.Hijacker")
}

// StatusCode returns the status code.
func (rw *responseWriterWrapper) StatusCode() int {
	return rw.statusCode
}

// Body returns the response body.
func (rw *responseWriterWrapper) Body() []byte {
	return rw.body
}
