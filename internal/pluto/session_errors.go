package pluto

import (
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
)

// SessionError preserves the application envelope for shared auth middleware.
func SessionError(
	w http.ResponseWriter,
	_ *http.Request,
	status int,
	code string,
) {
	JSON(w, status, api.ErrorResponse{Error: api.Error{
		Code:      code,
		Message:   "session authentication failed",
		RequestID: w.Header().Get("X-Request-ID"),
	}})
}
