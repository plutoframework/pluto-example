package app

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/plutoframework/pluto-example/internal/auth"
	"github.com/plutoframework/pluto-example/internal/config"

	"github.com/go-jose/go-jose/v4"
	"github.com/plutoframework/pluto-auth/httpauth"
	"github.com/plutoframework/pluto-auth/methods/password"
)

func TestNativeOIDCApplication(t *testing.T) {
	_, databaseURL := localDatabase(t)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: key},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	var issuer, nonce, challenge string

	provider := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch r.URL.Path {
			case "/.well-known/openid-configuration":
				_ = json.NewEncoder(w).
					Encode(map[string]any{"issuer": issuer, "authorization_endpoint": issuer + "/authorize", "token_endpoint": issuer + "/token", "jwks_uri": issuer + "/keys", "id_token_signing_alg_values_supported": []string{"RS256"}})
			case "/keys":
				_ = json.NewEncoder(w).
					Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, Algorithm: "RS256", Use: "sig"}}})
			case "/token":
				if err := r.ParseForm(); err != nil {
					t.Error(err)
				}

				hash := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
				if base64.RawURLEncoding.EncodeToString(hash[:]) != challenge {
					t.Error("missing PKCE")
				}

				payload, _ := json.Marshal(
					map[string]any{
						"iss":   issuer,
						"sub":   "oidc-example-user",
						"aud":   "example-client",
						"exp":   time.Now().Add(time.Minute).Unix(),
						"iat":   time.Now().Unix(),
						"nonce": nonce,
					},
				)

				signed, err := signer.Sign(payload)
				if err != nil {
					t.Error(err)

					return
				}

				token, err := signed.CompactSerialize()
				if err != nil {
					t.Error(err)

					return
				}

				_ = json.NewEncoder(w).
					Encode(map[string]string{"id_token": token, "access_token": "private", "token_type": "Bearer"})
			default:
				http.NotFound(w, r)
			}
		}),
	)
	defer provider.Close()

	issuer = provider.URL

	settings, err := config.LoadFrom(func(name string) (string, bool) {
		values := map[string]string{
			config.Prefix + "DEMO_AUTH":    "true",
			config.Prefix + "DATABASE_URL": databaseURL,
		}
		value, ok := values[name]

		return value, ok
	})
	if err != nil {
		t.Fatal(err)
	}

	encryption := make([]byte, 32)
	_, _ = rand.Read(encryption)
	settings.Auth.TrustedOrigins = []string{"https://app.example"}
	settings.OIDC = config.OIDC{
		Enabled:         true,
		AllowEnrollment: true,
		Issuer:          issuer,
		ClientID:        "example-client",
		ClientSecret:    "synthetic-only",
		CallbackURL:     "https://app.example/auth/oidc/callback",
		EncryptionKey:   encryption,
	}

	limiter, err := password.NewWindowLimiter(1000, 100, time.Minute, time.Now)
	if err != nil {
		t.Fatal(err)
	}

	ingress, err := auth.DevelopmentIngress()
	if err != nil {
		t.Fatal(err)
	}

	client := provider.Client()
	client.Timeout = 5 * time.Second

	application, err := NewWithLocal(
		context.Background(),
		settings,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		auth.LocalDependencies{
			Sender:         &deliveryRecorder{},
			Limiter:        limiter,
			Ingress:        ingress,
			OIDCHTTPClient: client,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer application.Close()

	call := func(method, path string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "https://app.example"+path, nil)
		r.RemoteAddr = "127.0.0.1:12345"
		r.Header.Set("Origin", "https://app.example")

		if method == "POST" {
			r.Header.Set("Accept", "application/json")
		}

		r.Header.Set("X-Organization-ID", "northstar")

		for _, cookie := range cookies {
			r.AddCookie(cookie)
		}

		response := httptest.NewRecorder()
		application.Handler.ServeHTTP(response, r)

		return response
	}

	available := call("GET", "/auth/providers")
	if available.Code != 200 ||
		available.Body.String() != "{\"data\":[\"oidc\"]}\n" {
		t.Fatal("provider not advertised")
	}

	begin := call("POST", "/auth/oidc/begin")
	if begin.Code != 200 {
		t.Fatalf("begin %d %s", begin.Code, begin.Body.String())
	}

	var beginBody struct {
		Data struct {
			URL string `json:"authorization_url"`
		} `json:"data"`
	}
	if json.Unmarshal(begin.Body.Bytes(), &beginBody) != nil {
		t.Fatal("invalid begin response")
	}

	parsed, _ := url.Parse(beginBody.Data.URL)
	query := parsed.Query()
	nonce = query.Get("nonce")
	challenge = query.Get("code_challenge")
	binding := begin.Result().Cookies()[0]

	callback := call(
		"GET",
		"/auth/oidc/callback?state="+query.Get("state")+"&code=synthetic",
		binding,
	)
	if callback.Code != 303 {
		t.Fatalf("callback %d %s", callback.Code, callback.Body.String())
	}

	var sessionCookie *http.Cookie

	for _, cookie := range callback.Result().Cookies() {
		if cookie.Name == httpauth.CookieName {
			sessionCookie = cookie
		}
	}

	if sessionCookie == nil {
		t.Fatal("no secure account session")
	}

	session := call("GET", "/auth/session", sessionCookie)
	if session.Code != 200 {
		t.Fatalf("session %d", session.Code)
	}

	var body struct {
		Data struct {
			ActorID string `json:"actor_id"`
		} `json:"data"`
	}
	if json.Unmarshal(session.Body.Bytes(), &body) != nil ||
		body.Data.ActorID == "" {
		t.Fatal("missing actor")
	}

	business := call("GET", "/api/purchases", sessionCookie)
	if business.Code == 200 {
		t.Fatal("sign-in auto-granted business membership")
	}

	replay := call(
		"GET",
		"/auth/oidc/callback?state="+query.Get("state")+"&code=synthetic",
		binding,
	)
	if replay.Code != 401 {
		t.Fatal("callback replay")
	}

	if err := application.Accounts.Disable(
		context.Background(),
		body.Data.ActorID,
	); err != nil {
		t.Fatal(err)
	}

	if call("GET", "/auth/session", sessionCookie).Code != 401 {
		t.Fatal("disabled session remains valid")
	}
}
