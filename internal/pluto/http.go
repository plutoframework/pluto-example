// Package pluto contains provisional transport helpers to review before extraction.
package pluto

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/domain"

	"github.com/plutoframework/pluto-core/pagination"
)

type contextKey int

const (
	principalKey contextKey = iota
	loggerKey
	requestIDKey
)

func Principal(request *http.Request) domain.Principal {
	principal, _ := request.Context().Value(principalKey).(domain.Principal)
	return principal
}

func Logger(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}

	return slog.New(slog.DiscardHandler)
}

func Decode(
	writer http.ResponseWriter,
	request *http.Request,
	target any,
	limit int64,
) error {
	request.Body = http.MaxBytesReader(writer, request.Body, limit)
	decoder := json.NewDecoder(request.Body)

	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return err
		}

		return domain.ErrInvalid
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return domain.ErrInvalid
	}

	// encoding/json normally accepts case-insensitive names. This adapter requires
	// the exact wire names advertised by the shared models and generated schema.
	model := reflect.TypeOf(target).Elem()
	allowed := make(map[string]bool)

	for index := 0; index < model.NumField(); index++ {
		name := strings.Split(model.Field(index).Tag.Get("json"), ",")[0]
		allowed[name] = true
	}

	for name := range fields {
		if !allowed[name] {
			return domain.ErrInvalid
		}
	}

	if err := json.Unmarshal(raw, target); err != nil {
		return err
	}

	return nil
}

func JSON(writer http.ResponseWriter, status int, body any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)

	// Wire models contain only JSON-encodable types. Connection failures cannot be
	// repaired after response headers have been sent.
	_ = json.NewEncoder(writer).Encode(body)
}

func Error(writer http.ResponseWriter, request *http.Request, err error) {
	status, code, message := http.StatusInternalServerError, "internal_error", "an unexpected error occurred"

	var (
		oversized *http.MaxBytesError
		syntax    *json.SyntaxError
		mismatch  *json.UnmarshalTypeError
	)
	switch {
	case errors.As(err, &oversized):
		status, code, message = 413, "body_too_large", "request body exceeds the limit"
	case errors.Is(err, domain.ErrUnauthenticated):
		status, code, message = 401, "unauthenticated", "valid session and organization membership required"
	case errors.Is(err, domain.ErrForbidden):
		status, code, message = 403, "forbidden", "permission or resource policy denied this action"
	case errors.Is(err, domain.ErrFeatureDisabled):
		status, code, message = 403, "feature_disabled", "this capability is disabled"
	case errors.Is(err, domain.ErrNotFound):
		status, code, message = 404, "not_found", "resource not found"
	case errors.Is(err, domain.ErrConflict):
		status, code, message = 409, "conflict", "resource changed or transition is invalid"
	case errors.Is(err, pagination.ErrInvalid),
		errors.Is(err, domain.ErrInvalid),
		errors.Is(err, io.EOF),
		errors.Is(err, io.ErrUnexpectedEOF),
		errors.As(err, &syntax),
		errors.As(err, &mismatch):
		status, code, message = 400, "invalid_input", "request fields or body are invalid"
	}

	if status == 500 {
		Logger(request.Context()).Error("request failed", "error", err)
	}

	if status == 401 {
		writer.Header().Set("WWW-Authenticate", `Bearer realm="pluto-example"`)
	}

	requestID, _ := request.Context().Value(requestIDKey).(string)
	JSON(writer, status, api.ErrorResponse{Error: api.Error{
		Code:      code,
		Message:   message,
		RequestID: requestID,
	}})
}

// DecodeJSON maps every decoder validation failure to the public invalid-input
// category while retaining body-limit errors for the 413 response.
func DecodeJSON(
	writer http.ResponseWriter,
	request *http.Request,
	target any,
	limit int64,
) error {
	err := Decode(writer, request, target, limit)
	if err == nil {
		return nil
	}

	var oversized *http.MaxBytesError
	if errors.As(err, &oversized) {
		return err
	}

	return domain.ErrInvalid
}
