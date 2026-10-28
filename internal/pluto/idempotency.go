package pluto

import (
	"net/http"

	"github.com/plutoframework/pluto-core/httpidem"
	"github.com/plutoframework/pluto-core/idempotency"
)

func NewIdempotency(
	store idempotency.Store,
	bodyLimit int64,
) (*httpidem.Handler, error) {
	return httpidem.New(
		store,
		idempotency.DefaultTTL,
		bodyLimit,
		256*1024,
		func(w http.ResponseWriter, r *http.Request, err error) {
			status := httpidem.Status(err)
			code := "idempotency_unavailable"

			switch status {
			case 400:
				code = "invalid_idempotency_key"
			case 409:
				code = "idempotency_conflict_or_pending"
			case 413:
				code = "body_too_large"
			}

			JSON(
				w,
				status,
				map[string]any{
					"error": map[string]any{
						"code":       code,
						"message":    http.StatusText(status),
						"request_id": w.Header().Get("X-Request-ID"),
					},
				},
			)
		},
	)
}
