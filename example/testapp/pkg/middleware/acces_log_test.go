package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type panicHandler struct{}

func (t panicHandler) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {
	panic("test_message")
}

type badRequestHandler struct{}

func (t badRequestHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte("bad_request"))
}

func TestAccessLogMiddleware_BadRequest(t *testing.T) {
	sink := &MemorySink{new(bytes.Buffer)}
	logger := tNewLogger(t, "alSink", sink)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(context.XRequestIDHeader, "test_request_id")
	r = r.WithContext(context.InitFromHTTP(w, r, ""))

	handler := badRequestHandler{}
	middleware := AccessLogMiddleware(logger)
	middleware(handler).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	expectedContains := []string{
		`"level":"info"`,
		`"msg":"access_log"`,
		`"method":"GET","path":"/"`,
		`"status_code":400`,
		`"duration":`,
		`"response_body":"bad_request"`,
		`"request_id":"test_request_id"`,
	}

	for _, expected := range expectedContains {
		assert.Contains(t, sink.String(), expected)
	}
}

func TestAccessLogMiddleware_PanicRecovery(t *testing.T) {
	sink := &MemorySink{new(bytes.Buffer)}
	logger := tNewLogger(t, "alPanicSink", sink)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(context.XRequestIDHeader, "test_request_id")
	r = r.WithContext(context.InitFromHTTP(w, r, ""))

	handler := panicHandler{}
	middleware := AccessLogMiddleware(logger)
	middleware(handler).ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	expectedContains := []string{
		`"level":"error"`,
		`"msg":"access_log"`,
		`"method":"GET","path":"/"`,
		`"event":"recovered after panic"`,
		`"panic_value":"test_message"`,
		`"status_code":500`,
		`"stacktrace":"goroutine`,
		`"request_id":"test_request_id"`,
	}

	for _, expected := range expectedContains {
		assert.Contains(t, sink.String(), expected)
	}
}

func tNewLogger(t *testing.T, sinkName string, sink zap.Sink) zlog.Logger {
	err := zap.RegisterSink(sinkName, func(*url.URL) (zap.Sink, error) {
		return sink, nil
	})
	require.NoError(t, err)

	logger := zlog.Nop()
	require.NoError(t, err)

	return logger
}

// MemorySink implements zap.Sink by writing all messages to a buffer.
type MemorySink struct {
	*bytes.Buffer
}

// Implement Close and Sync as no-ops to satisfy the interface. The Write
// method is provided by the embedded buffer.
func (s *MemorySink) Close() error { return nil }
func (s *MemorySink) Sync() error  { return nil }
