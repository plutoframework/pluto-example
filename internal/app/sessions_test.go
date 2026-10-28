package app

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/plutoframework/pluto-example/internal/config"

	"github.com/plutoframework/pluto-auth/httpauth"
)

func TestSessionLoginMembershipAndLogout(t *testing.T) {
	settings, err := config.LoadFrom(
		func(key string) (string, bool) { return "true", key == config.Prefix+"DEMO_AUTH" },
	)
	if err != nil {
		t.Fatal(err)
	}

	application, err := New(
		context.Background(),
		settings,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer application.Close()

	request := func(method, path, token, tenant, origin string, cookie bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, nil)
		if cookie {
			r.AddCookie(&http.Cookie{Name: httpauth.CookieName, Value: token})
		} else if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}

		if tenant != "" {
			r.Header.Set("X-Organization-ID", tenant)
		}

		if origin != "" {
			r.Header.Set("Origin", origin)
		}

		w := httptest.NewRecorder()
		application.Handler.ServeHTTP(w, r)

		return w
	}

	const (
		credential = "demo_bob_tW4nC6yP1sF3"
		origin     = "http://127.0.0.1:8080"
	)

	if got := request(
		"GET",
		"/api/purchases",
		credential,
		"northstar",
		"",
		false,
	).Code; got != 401 {
		t.Fatal("bootstrap credential reached business API")
	}

	if got := request(
		"POST",
		"/auth/session",
		credential,
		"",
		"",
		false,
	).Code; got != 403 {
		t.Fatal("login CSRF guard missing")
	}

	if got := request(
		"POST",
		"/auth/session",
		"invalid",
		"",
		origin,
		false,
	).Code; got != 401 {
		t.Fatal("invalid credential issued a session")
	}

	login := request("POST", "/auth/session", credential, "", origin, false)
	if login.Code != 201 {
		t.Fatalf("login=%d %s", login.Code, login.Body)
	}

	cookies := login.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].Secure || !cookies[0].HttpOnly ||
		cookies[0].Domain != "" {
		t.Fatal("secure session cookie missing")
	}

	token := cookies[0].Value
	if strings.Contains(login.Body.String(), token) ||
		strings.Contains(login.Body.String(), credential) {
		t.Fatal("credential leaked through response JSON")
	}

	for _, test := range []struct {
		method, path, tenant, origin string
		cookie                       bool
		want                         int
	}{
		{"GET", "/api/purchases", "northstar", "", true, 200},
		{"GET", "/api/purchases", "harbor", "", true, 401},
		{"GET", "/api/purchases", "", "", true, 401},
		{"POST", "/api/purchases", "northstar", "", true, 403},
		{"GET", "/api/purchases", "northstar", "", false, 200},
		{"GET", "/auth/session", "", "", true, 200},
		{"DELETE", "/auth/session", "", "", true, 403},
		{"DELETE", "/auth/session", "", origin, true, 200},
		{"GET", "/api/purchases", "northstar", "", true, 401},
		{"GET", "/auth/session", "", "", false, 401},
	} {
		response := request(
			test.method,
			test.path,
			token,
			test.tenant,
			test.origin,
			test.cookie,
		)
		if response.Code != test.want {
			t.Fatalf(
				"%s %s status=%d want=%d body=%s",
				test.method,
				test.path,
				response.Code,
				test.want,
				response.Body,
			)
		}

		assertSecurityHeaders(t, response)

		if test.want >= 400 {
			var body struct {
				Error struct {
					RequestID string `json:"request_id"`
				} `json:"error"`
			}
			if err := json.Unmarshal(
				response.Body.Bytes(),
				&body,
			); err != nil ||
				body.Error.RequestID == "" {
				t.Fatal("auth error lost envelope/request ID")
			}
		}
	}
}
