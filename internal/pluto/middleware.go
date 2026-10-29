package pluto

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

type requestState struct {
	logger *slog.Logger
}

const stateKey contextKey = 100

// Middleware generates its own correlation ID and logs verified identity only.
func Middleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				var id [12]byte
				if _, err := rand.Read(id[:]); err != nil {
					http.Error(
						writer,
						"request ID unavailable",
						http.StatusInternalServerError,
					)

					return
				}

				requestID := hex.EncodeToString(id[:])
				writer.Header().Set("X-Request-ID", requestID)

				state := &requestState{
					logger: logger.With(
						"request_id",
						requestID,
						"method",
						request.Method,
					),
				}
				ctx := context.WithValue(
					request.Context(),
					requestIDKey,
					requestID,
				)
				ctx = context.WithValue(ctx, loggerKey, state.logger)
				ctx = context.WithValue(ctx, stateKey, state)
				request = request.WithContext(ctx)
				response := &statusWriter{ResponseWriter: writer}
				started := time.Now()

				defer func() {
					if recovered := recover(); recovered != nil {
						// Panic values may contain application data. Do not log them.
						state.logger.ErrorContext(ctx, "request panic")

						if response.status == 0 {
							Error(response, request, errPanic{})
						}
					}

					status := response.status
					if status == 0 {
						status = http.StatusOK
					}

					state.logger.InfoContext(
						ctx,
						"request completed",
						"status",
						status,
						"duration_ms",
						time.Since(started).Milliseconds(),
					)
				}()

				next.ServeHTTP(response, request)
			},
		)
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}

	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}

	return w.ResponseWriter.Write(body)
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

type errPanic struct{}

func (errPanic) Error() string { return "request panic" }
