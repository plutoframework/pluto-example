package pluto

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type correlationKey struct{}

type contextHandler struct {
	slog.Handler
	contexts []context.Context
}

func (h *contextHandler) Enabled(
	context.Context,
	slog.Level,
) bool {
	return true
}

func (h *contextHandler) Handle(ctx context.Context, _ slog.Record) error {
	h.contexts = append(h.contexts, ctx)
	return nil
}
func (h *contextHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *contextHandler) WithGroup(string) slog.Handler      { return h }

// Trace correlation relies on slog receiving the request context, including on
// recovered panics and the completion record that follows them.
func TestRequestLogsRetainContext(t *testing.T) {
	for _, panicRequest := range []bool{false, true} {
		handler := &contextHandler{}
		next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			if panicRequest {
				panic("private data")
			}
		})
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request = request.WithContext(
			context.WithValue(
				request.Context(),
				correlationKey{},
				"trace-context",
			),
		)
		Middleware(
			slog.New(handler),
		)(
			next,
		).ServeHTTP(httptest.NewRecorder(), request)

		expected := 1
		if panicRequest {
			expected = 3
		}

		if len(handler.contexts) != expected {
			t.Fatalf("got %d records, want %d", len(handler.contexts), expected)
		}

		for _, ctx := range handler.contexts {
			if ctx.Value(correlationKey{}) != "trace-context" {
				t.Fatal("request log lost trace context")
			}
		}
	}
}
