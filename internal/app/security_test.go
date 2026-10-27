package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/plutoframework/pluto-example/internal/paths"
)

func TestRouterSecurityHeaders(t *testing.T) {
	for _, test := range []struct {
		name     string
		endpoint paths.Endpoint
		token    string
		body     string
		status   int
	}{
		{
			name:     "success",
			endpoint: paths.ListPurchases,
			token:    aliceToken,
			status:   http.StatusOK,
		},
		{
			name:     "unauthenticated",
			endpoint: paths.ListPurchases,
			status:   http.StatusUnauthorized,
		},
		{
			name:     "forbidden",
			endpoint: paths.ExportPurchases,
			token:    aliceToken,
			status:   http.StatusForbidden,
		},
		{
			name:     "invalid JSON",
			endpoint: paths.CreatePurchase,
			token:    aliceToken,
			body:     "{",
			status:   http.StatusBadRequest,
		},
		{
			name:     "oversized body",
			endpoint: paths.CreatePurchase,
			token:    aliceToken,
			body:     `{"title":"` + strings.Repeat("x", 600) + `"}`,
			status:   http.StatusRequestEntityTooLarge,
		},
		{
			name: "missing route",
			endpoint: paths.Endpoint{
				Method: http.MethodGet,
				Path:   "/missing",
			},
			status: http.StatusNotFound,
		},
		{
			name: "method not allowed",
			endpoint: paths.Endpoint{
				Method: http.MethodDelete,
				Path:   paths.ListPurchases.Path,
			},
			status: http.StatusMethodNotAllowed,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := call(
				testRouter(t, &operationSpy{}, true, io.Discard),
				test.endpoint,
				test.token,
				"northstar",
				test.body,
			)
			if response.Code != test.status {
				t.Fatalf("status=%d body=%s", response.Code, response.Body)
			}

			assertSecurityHeaders(t, response)
		})
	}
}

func TestRouterSecurityHeadersOnPanic(t *testing.T) {
	// The spy's unimplemented Get intentionally panics through its nil interface.
	endpoint := paths.GetPurchase
	endpoint.Path = strings.ReplaceAll(endpoint.Path, "{purchaseID}", "test-id")

	response := call(
		testRouter(t, &operationSpy{}, true, io.Discard),
		endpoint,
		aliceToken,
		"northstar",
		"",
	)
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("panic status=%d body=%s", response.Code, response.Body)
	}

	assertSecurityHeaders(t, response)
}

func assertSecurityHeaders(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()

	for name, want := range map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "no-referrer",
		"Cache-Control":          "no-store",
	} {
		if got := response.Result().Header.Get(name); got != want {
			t.Errorf("%s=%q, want %q", name, got, want)
		}
	}

	if response.Header().Get("Content-Security-Policy") == "" {
		t.Error("missing Content-Security-Policy")
	}
}
