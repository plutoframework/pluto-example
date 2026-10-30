package app

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/plutoframework/pluto-example/internal/auth"
	"github.com/plutoframework/pluto-example/internal/config"
	"github.com/plutoframework/pluto-example/internal/development"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/plutoframework/pluto-auth/httpauth"
	"github.com/plutoframework/pluto-auth/methods/password"
)

// Bounded in-process test delivery only. Never logs or persists notifications.
type deliveryRecorder struct {
	mu       sync.Mutex
	items    []password.Notification
	fail     bool
	failures int
}

func (s *deliveryRecorder) Send(
	ctx context.Context,
	n password.Notification,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.fail || len(s.items) >= 40 {
		return errors.New("sensitive provider failure")
	}

	s.items = append(s.items, n)

	return nil
}

func (s *deliveryRecorder) token(
	t *testing.T,
	email string,
	purpose password.Purpose,
) string {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := len(s.items) - 1; i >= 0; i-- {
		if s.items[i].Email == email && s.items[i].Purpose == purpose {
			return s.items[i].Token
		}
	}

	t.Fatal("notification missing")

	return ""
}

func localDatabase(t *testing.T) (*pgxpool.Pool, string) {
	t.Helper()

	raw := os.Getenv("PLUTO_EXAMPLE_TEST_DATABASE_URL")
	if raw == "" {
		t.Skip("requires disposable PostgreSQL")
	}

	ctx := context.Background()

	admin, err := pgxpool.New(ctx, raw)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(admin.Close)

	schema := "lifecycle_" + strings.ToLower(rand.Text())

	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+identifier); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if _, err := admin.Exec(
			cleanup,
			"DROP SCHEMA "+identifier+" CASCADE",
		); err != nil {
			t.Error("schema cleanup failed")
		}
	})

	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}

	query := parsed.Query()
	query.Set("search_path", schema)

	parsed.RawQuery = query.Encode()

	pool, err := pgxpool.New(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(pool.Close)

	for _, name := range []string{"migrations/001_initial.sql", "migrations/002_auth_sessions.sql", "migrations/003_idempotency.sql", "migrations/004_local_credentials.sql", "migrations/005_jobs.sql", "migrations/006_authentication.sql", "fixtures.sql"} {
		migration, err := os.ReadFile(filepath.Join("../../db", name))
		if err != nil {
			t.Fatal(err)
		}

		if _, err := pool.Exec(ctx, string(migration)); err != nil {
			t.Fatal(err)
		}
	}

	return pool, parsed.String()
}

type lifecycleApp struct {
	accounts auth.AccountDisabler
	handler  http.Handler
	delivery *deliveryRecorder
	logs     *bytes.Buffer
	pool     *pgxpool.Pool
}

func newLifecycleApp(t *testing.T) lifecycleApp {
	t.Helper()

	pool, databaseURL := localDatabase(t)

	settings, err := config.LoadFrom(func(key string) (string, bool) {
		cacheEnvironment := sha256.Sum256([]byte(databaseURL))

		values := map[string]string{
			config.Prefix + "DEMO_AUTH":    "true",
			config.Prefix + "DATABASE_URL": databaseURL,
			config.Prefix + "CACHE_ENVIRONMENT": hex.EncodeToString(
				cacheEnvironment[:],
			),
		}
		if address := os.Getenv("PLUTO_EXAMPLE_CACHE_TEST_URL"); address != "" {
			values[config.Prefix+"CACHE_BACKEND"] = "valkey"
			values[config.Prefix+"CACHE_URL"] = address
		}

		value, ok := values[key]

		return value, ok
	})
	if err != nil {
		t.Fatal(err)
	}

	delivery := &deliveryRecorder{}

	limiter, err := password.NewWindowLimiter(1000, 100, time.Minute, time.Now)
	if err != nil {
		t.Fatal(err)
	}

	ingress, err := auth.DevelopmentIngress()
	if err != nil {
		t.Fatal(err)
	}

	logs := &bytes.Buffer{}

	application, err := NewWithLocal(context.Background(), settings,
		slog.New(slog.NewJSONHandler(logs, nil)), auth.LocalDependencies{
			Sender: delivery, Limiter: limiter, Ingress: ingress,
			DeliveryError: func(_ context.Context, err error) {
				if !errors.Is(err, password.ErrDelivery) {
					t.Error("unredacted delivery error")
				}

				delivery.failures++
			},
		})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(application.Close)

	return lifecycleApp{
		accounts: application.Accounts,
		handler:  application.Handler,
		delivery: delivery,
		logs:     logs,
		pool:     pool,
	}
}

const (
	testOrigin   = "http://127.0.0.1:8080"
	testPassword = "synthetic test passphrase only"
)

func (a lifecycleApp) request(
	t *testing.T,
	method, path, token, tenant, origin string,
	input any,
	want int,
) *httptest.ResponseRecorder {
	t.Helper()

	var body io.Reader

	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		body = bytes.NewReader(encoded)
	}

	r := httptest.NewRequest(method, path, body)
	r.RemoteAddr = "127.0.0.1:12345"
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Origin", origin)
	r.Header.Set("X-Organization-ID", tenant)
	// Identical key on every auth request MUST NOT cache tokens/cookies/results.
	r.Header.Set("Idempotency-Key", "lifecycle-request")

	if token != "" {
		r.AddCookie(&http.Cookie{Name: httpauth.CookieName, Value: token})
	}

	w := httptest.NewRecorder()
	a.handler.ServeHTTP(w, r)

	if w.Code != want {
		t.Fatalf("%s %s status=%d want=%d", method, path, w.Code, want)
	}

	assertSecurityHeaders(t, w)

	return w
}

func (a lifecycleApp) post(
	t *testing.T,
	path string,
	input any,
	want int,
) *httptest.ResponseRecorder {
	t.Helper()

	return a.request(t, "POST", path, "", "", testOrigin, input, want)
}

func (a lifecycleApp) signin(
	t *testing.T,
	email, value string,
	want int,
) string {
	t.Helper()

	w := a.post(
		t,
		"/auth/signin",
		map[string]string{"email": email, "password": value},
		want,
	)
	if want != 200 {
		return ""
	}

	cookies := w.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].Secure || !cookies[0].HttpOnly ||
		cookies[0].Domain != "" {
		t.Fatal("cookie contract")
	}

	if strings.Contains(w.Body.String(), cookies[0].Value) {
		t.Fatal("session leaked")
	}

	var recorded bool

	err := a.pool.QueryRow(context.Background(), `SELECT EXISTS (
        SELECT 1 FROM pluto_auth_local_sessions s
        JOIN pluto_auth_accounts a ON a.id = s.account_id
        WHERE a.email = $1 AND s.primary_method = 'password'
            AND s.primary_at > '0001-01-01 00:00:00+00'
    )`, email).Scan(&recorded)
	if err != nil || !recorded {
		t.Fatal("sign-in did not persist password authentication evidence", err)
	}

	return cookies[0].Value
}

func (a lifecycleApp) enroll(t *testing.T, email string) string {
	t.Helper()
	a.post(
		t,
		"/auth/enroll",
		map[string]string{"email": email, "password": testPassword},
		202,
	)
	a.signin(t, email, testPassword, 401)

	token := a.delivery.token(t, email, password.VerifyEmail)
	a.post(t, "/auth/verify", map[string]string{"token": token}, 200)
	a.post(t, "/auth/verify", map[string]string{"token": token}, 400)

	return a.signin(t, email, testPassword, 200)
}

func TestPostgresLocalLifecycle(t *testing.T) {
	a := newLifecycleApp(t)
	a.signin(t, "unknown@example.test", testPassword, 401)

	subjects := map[string]string{}
	cookies := map[string]string{}

	for _, name := range []string{"alice", "bob", "carol", "dana"} {
		cookie := a.enroll(t, name+"@example.test")
		cookies[name] = cookie
		a.request(t, "GET", "/api/purchases", cookie, "northstar", "", nil, 401)

		w := a.request(t, "GET", "/auth/session", cookie, "", "", nil, 200)

		var result struct {
			Data struct {
				ActorID string `json:"actor_id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}

		subjects[name] = result.Data.ActorID
	}

	if err := development.BindFixtures(
		context.Background(),
		a.pool,
		subjects,
	); err != nil {
		t.Fatal(err)
	}

	alice := cookies["alice"]
	a.request(
		t,
		"GET",
		"/api/purchases?limit=1",
		alice,
		"northstar",
		"",
		nil,
		200,
	)
	a.request(
		t,
		"GET",
		"/api/purchases",
		cookies["bob"],
		"harbor",
		"",
		nil,
		401,
	)

	input := map[string]any{"title": "Microscope", "amount_cents": 100}
	first := a.request(
		t,
		"POST",
		"/api/purchases",
		alice,
		"northstar",
		testOrigin,
		input,
		201,
	)

	replay := a.request(
		t,
		"POST",
		"/api/purchases",
		alice,
		"northstar",
		testOrigin,
		input,
		201,
	)
	if first.Body.String() != replay.Body.String() {
		t.Fatal("business replay changed")
	}

	a.post(
		t,
		"/auth/recovery",
		map[string]string{"email": "alice@example.test"},
		202,
	)

	reset := a.delivery.token(t, "alice@example.test", password.RecoverPassword)
	a.post(
		t,
		"/auth/reset",
		map[string]string{
			"token":    reset,
			"password": "replacement test passphrase",
		},
		200,
	)
	a.post(
		t,
		"/auth/reset",
		map[string]string{"token": reset, "password": testPassword},
		400,
	)
	a.request(t, "GET", "/api/purchases", alice, "northstar", "", nil, 401)
	a.request(
		t,
		"POST",
		"/api/purchases",
		alice,
		"northstar",
		testOrigin,
		input,
		401,
	)
	a.signin(t, "alice@example.test", testPassword, 401)

	alice = a.signin(
		t,
		"alice@example.test",
		"replacement test passphrase",
		200,
	)

	second := a.signin(
		t,
		"alice@example.test",
		"replacement test passphrase",
		200,
	)
	if alice == second {
		t.Fatal("signin cookie was replayed")
	}

	a.request(t, "POST", "/auth/logout", alice, "", "", nil, 403)
	a.request(t, "POST", "/auth/logout", alice, "", testOrigin, nil, 200)
	a.request(t, "GET", "/auth/session", alice, "", "", nil, 401)

	if err := a.accounts.Disable(
		context.Background(),
		subjects["alice"],
	); err != nil {
		t.Fatal(err)
	}

	a.request(t, "GET", "/api/purchases", second, "northstar", "", nil, 401)
	a.signin(t, "alice@example.test", "replacement test passphrase", 401)

	for _, n := range a.delivery.items {
		if strings.Contains(a.logs.String(), n.Token) {
			t.Fatal("challenge logged")
		}
	}

	for _, secret := range []string{testPassword, reset, alice, second, "replacement test passphrase"} {
		if strings.Contains(a.logs.String(), secret) {
			t.Fatal("credential logged")
		}
	}
}

func TestPostgresLocalDeliveryResendAndOrigin(t *testing.T) {
	a := newLifecycleApp(t)
	input := map[string]string{
		"email":    "new@example.test",
		"password": testPassword,
	}
	a.request(t, "POST", "/auth/enroll", "", "", "", input, 403)
	a.request(
		t,
		"POST",
		"/auth/enroll",
		"",
		"",
		"https://evil.example.test",
		input,
		403,
	)
	a.request(t, "GET", "/auth/verify", "", "", "", nil, 405)

	a.delivery.fail = true
	a.post(t, "/auth/enroll", input, 202)

	if a.delivery.failures != 1 {
		t.Fatal("delivery failure not observed")
	}

	a.delivery.fail = false
	a.post(t, "/auth/enroll", input, 202)

	if len(a.delivery.items) != 0 {
		t.Fatal("duplicate enrollment resent")
	}

	a.post(
		t,
		"/auth/verification",
		map[string]string{"email": "new@example.test"},
		202,
	)

	old := a.delivery.token(t, "new@example.test", password.VerifyEmail)
	a.post(
		t,
		"/auth/verification",
		map[string]string{"email": "new@example.test"},
		202,
	)

	latest := a.delivery.token(t, "new@example.test", password.VerifyEmail)
	a.post(t, "/auth/verify", map[string]string{"token": old}, 400)
	a.post(
		t,
		"/auth/reset",
		map[string]string{"token": latest, "password": testPassword},
		400,
	)
	a.post(t, "/auth/verify", map[string]string{"token": latest}, 200)

	cookie := a.signin(t, "new@example.test", testPassword, 200)
	a.request(t, "GET", "/api/purchases", cookie, "northstar", "", nil, 401)
	a.post(t, "/auth/verify", map[string]string{"token": "invalid"}, 400)
	a.post(
		t,
		"/auth/enroll",
		map[string]string{
			"email":    "new@example.test",
			"password": testPassword,
			"role":     "administrator",
		},
		400,
	)

	var count int
	if err := a.pool.QueryRow(context.Background(), "SELECT count(*) FROM memberships WHERE actor_id IN (SELECT id FROM pluto_auth_accounts)").
		Scan(&count); err != nil ||
		count != 0 {
		t.Fatal("automatic membership grant")
	}
}
