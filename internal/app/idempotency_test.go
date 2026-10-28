package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBusinessReplayScopeConflictAndAuthentication(t *testing.T) {
	spy := &operationSpy{}
	router := testRouter(t, spy, false, io.Discard)
	send := func(token, tenant, key, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(
			"POST",
			"/api/purchases",
			strings.NewReader(body),
		)
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("X-Organization-ID", tenant)
		r.Header.Set("Idempotency-Key", key)
		r.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		return w
	}
	first := send(
		aliceToken,
		"northstar",
		"retry-create",
		`{"title":"Equipment","amount_cents":100}`,
	)

	replay := send(
		aliceToken,
		"northstar",
		"retry-create",
		`{"title":"Equipment","amount_cents":100}`,
	)
	if first.Code != 201 || replay.Code != 201 ||
		first.Body.String() != replay.Body.String() ||
		spy.calls != 1 {
		t.Fatal(first, replay, spy.calls)
	}

	if w := send(
		aliceToken,
		"northstar",
		"retry-create",
		`{"title":"Changed","amount_cents":100}`,
	); w.Code != 409 ||
		spy.calls != 1 {
		t.Fatal(w, spy.calls)
	}

	if w := send(
		"invalid",
		"northstar",
		"retry-create",
		`{"title":"Equipment","amount_cents":100}`,
	); w.Code != http.StatusUnauthorized ||
		spy.calls != 1 {
		t.Fatal("replay bypassed auth", w, spy.calls)
	}

	if w := send(
		aliceToken,
		"harbor",
		"retry-create",
		`{"title":"Equipment","amount_cents":100}`,
	); w.Code != 201 ||
		spy.calls != 2 {
		t.Fatal("tenant keys collided", w, spy.calls)
	}

	if w := send(
		aliceToken,
		"northstar",
		"",
		`{"title":"Equipment","amount_cents":100}`,
	); w.Code != 400 ||
		spy.calls != 2 {
		t.Fatal("missing client key accepted", w, spy.calls)
	}
}
