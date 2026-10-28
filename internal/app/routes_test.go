package app

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/auth"
	"github.com/plutoframework/pluto-example/internal/config"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/features"
	"github.com/plutoframework/pluto-example/internal/paths"
	"github.com/plutoframework/pluto-example/internal/pluto"
	"github.com/plutoframework/pluto-example/internal/storage/memory"

	plutoauth "github.com/plutoframework/pluto-auth"
	"github.com/plutoframework/pluto-auth/httpauth"
	authmemory "github.com/plutoframework/pluto-auth/stores/memory"
)

type operationSpy struct {
	ApplicationOperations
	calls     int
	principal domain.Principal
}

func (s *operationSpy) Create(
	_ context.Context,
	principal domain.Principal,
	title string,
	cents int64,
) (domain.Purchase, error) {
	s.calls++
	s.principal = principal

	return domain.Purchase{
		ID:          "created",
		TenantID:    principal.TenantID,
		CreatorID:   principal.ActorID,
		Title:       title,
		AmountCents: cents,
		Status:      domain.Draft,
	}, nil
}

func (s *operationSpy) Export(
	_ context.Context,
	principal domain.Principal,
) ([]domain.Purchase, error) {
	s.calls++
	s.principal = principal

	return []domain.Purchase{}, nil
}

func (s *operationSpy) List(
	_ context.Context,
	principal domain.Principal,
) ([]domain.Purchase, error) {
	s.calls++
	s.principal = principal

	return []domain.Purchase{{
		ID:    "request",
		Title: strings.Repeat("x", 150),
	}}, nil
}

const (
	aliceToken = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	bobToken   = "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE"
)

func testRouter(
	t testing.TB,
	spy *operationSpy,
	enabled bool,
	logWriter io.Writer,
) http.Handler {
	t.Helper()

	policy, err := domain.NewPolicy()
	if err != nil {
		t.Fatal(err)
	}

	settings := config.ServerConfig{
		MaxBodyBytes: 512,
		Gzip:         true,
	}
	logger := slog.New(slog.NewJSONHandler(logWriter, nil))
	store := memory.New(auth.DemoMemberships())

	sessions, err := authmemory.New(100)
	if err != nil {
		t.Fatal(err)
	}

	for token, actor := range map[string]string{aliceToken: "alice", bobToken: "bob"} {
		digest, err := plutoauth.TokenDigest(token)
		if err != nil {
			t.Fatal(err)
		}

		now := time.Now().UTC()

		err = sessions.Create(context.Background(), digest, plutoauth.Session{
			ID:         actor,
			SubjectID:  actor,
			CreatedAt:  now,
			LastSeenAt: now,
			ExpiresAt:  now.Add(time.Hour),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	manager, err := plutoauth.New(sessions, plutoauth.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	browser, err := httpauth.New(
		manager,
		httpauth.Config{
			TrustedOrigins: []string{"https://app.example.com"},
			AllowBearer:    true,
			WriteError:     pluto.SessionError,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	return Routes(
		settings,
		logger,
		SessionAuthentication{
			Resolver: auth.Sessions{Memberships: store},
			HTTP:     browser,
		},
		features.NewSnapshot(
			map[features.Key]bool{features.PurchaseExport: enabled},
		),
		spy,
		policy,
	)
}

func call(
	handler http.Handler,
	endpoint paths.Endpoint,
	token, tenant, body string,
) *httptest.ResponseRecorder {
	request := httptest.NewRequest(
		endpoint.Method,
		endpoint.Path,
		strings.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")

	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	if tenant != "" {
		request.Header.Set("X-Organization-ID", tenant)
	}

	writer := httptest.NewRecorder()
	handler.ServeHTTP(writer, request)

	return writer
}

func TestDeniedRequestsNeverInvokeOperations(t *testing.T) {
	for _, test := range []struct {
		name, token, tenant string
		enabled             bool
		status              int
		code                string
	}{
		{
			name:   "no auth",
			tenant: "northstar",
			status: 401,
			code:   "unauthenticated",
		},
		{
			name:   "invalid token",
			token:  "forged",
			tenant: "northstar",
			status: 401,
			code:   "unauthenticated",
		},
		{
			name:   "no tenant",
			token:  bobToken,
			status: 401,
			code:   "unauthenticated",
		},
		{
			name:   "unauthorized tenant",
			token:  bobToken,
			tenant: "harbor",
			status: 401,
			code:   "unauthenticated",
		},
		{
			name:   "flag disabled",
			token:  bobToken,
			tenant: "northstar",
			status: 403,
			code:   "feature_disabled",
		},
		{
			name:    "flag not permission",
			token:   aliceToken,
			tenant:  "northstar",
			enabled: true,
			status:  403,
			code:    "forbidden",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			spy := &operationSpy{}

			response := call(
				testRouter(t, spy, test.enabled, io.Discard),
				paths.ExportPurchases,
				test.token,
				test.tenant,
				"",
			)
			if response.Code != test.status || spy.calls != 0 {
				t.Fatalf(
					"status=%d calls=%d body=%s",
					response.Code,
					spy.calls,
					response.Body,
				)
			}

			var body api.ErrorResponse
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}

			if body.Error.Code != test.code || body.Error.RequestID == "" {
				t.Fatalf("bad error envelope: %+v", body)
			}

			if response.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("personalized response may be cached")
			}
		})
	}
}

func TestAdapterExtractsVerifiedPrincipal(t *testing.T) {
	spy := &operationSpy{}

	response := call(
		testRouter(t, spy, true, io.Discard),
		paths.CreatePurchase,
		aliceToken,
		"harbor",
		`{"title":"Safety equipment","amount_cents":5000}`,
	)
	if response.Code != 201 || spy.calls != 1 {
		t.Fatalf("status=%d body=%s", response.Code, response.Body)
	}

	if spy.principal.ActorID != "alice" || spy.principal.TenantID != "harbor" ||
		spy.principal.Role != domain.Member {
		t.Fatalf("bad principal: %+v", spy.principal)
	}

	var body api.PurchaseResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}

	if body.Data.Title != "Safety equipment" || body.Data.AmountCents != 5000 {
		t.Fatalf("wire mapping lost values: %+v", body)
	}
}

func TestInvalidBodiesNeverInvokeOperations(t *testing.T) {
	for _, test := range []struct {
		name, body string
		status     int
	}{
		{
			"unknown tenant field",
			`{"title":"x","amount_cents":1,"organization_id":"harbor"}`,
			400,
		},
		{
			"case mismatch",
			`{"Title":"x","AMOUNT_CENTS":1}`,
			400,
		},
		{
			"malformed",
			`{"title":`,
			400,
		},
		{
			"trailing JSON",
			`{} {}`,
			400,
		},
		{
			"wrong type",
			`{"amount_cents":"1"}`,
			400,
		},
		{
			"null",
			"null",
			400,
		},
		{
			"array",
			"[]",
			400,
		},
		{
			"oversized",
			fmt.Sprintf(`{"title":%q,"amount_cents":1}`, strings.Repeat("x", 600)),
			413,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			spy := &operationSpy{}

			response := call(
				testRouter(t, spy, false, io.Discard),
				paths.CreatePurchase,
				aliceToken,
				"northstar",
				test.body,
			)
			if response.Code != test.status || spy.calls != 0 {
				t.Fatalf(
					"status=%d calls=%d body=%s",
					response.Code,
					spy.calls,
					response.Body,
				)
			}
		})
	}
}

func TestVisibleFlagsUseSameEvaluatorAndRequireAuth(t *testing.T) {
	for _, enabled := range []bool{
		false,
		true,
	} {
		spy := &operationSpy{}
		handler := testRouter(t, spy, enabled, io.Discard)
		response := call(handler, paths.GetFeatures, bobToken, "northstar", "")

		var body api.FeaturesResponse
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}

		if response.Code != 200 || len(body.Data) != 1 ||
			body.Data["purchase_export"] != enabled {
			t.Fatalf("bad flags: status=%d %+v", response.Code, body)
		}

		export := call(
			handler,
			paths.ExportPurchases,
			bobToken,
			"northstar",
			"",
		)
		if enabled && (export.Code != 200 || spy.calls != 1) {
			t.Fatal("visible enabled flag disagrees with gate")
		}

		if !enabled && (export.Code != 403 || spy.calls != 0) {
			t.Fatal("visible disabled flag disagrees with gate")
		}
	}

	spy := &operationSpy{}
	if response := call(
		testRouter(t, spy, true, io.Discard),
		paths.GetFeatures,
		"",
		"northstar",
		"",
	); response.Code != 401 {
		t.Fatal("anonymous feature discovery allowed")
	}
}

func TestCompletionLogsAreEnrichedAndDoNotContainCredentials(t *testing.T) {
	var logs bytes.Buffer

	spy := &operationSpy{}
	response := call(
		testRouter(t, spy, false, &logs),
		paths.ExportPurchases,
		bobToken,
		"northstar",
		"",
	)

	var record map[string]any
	if err := json.Unmarshal(
		bytes.TrimSpace(logs.Bytes()),
		&record,
	); err != nil {
		t.Fatalf("bad diagnostic JSON: %s: %v", logs.String(), err)
	}

	if record["endpoint_id"] != "purchases.export" ||
		record["actor_id"] != "bob" ||
		record["tenant_id"] != "northstar" ||
		record["status"] != float64(403) {
		t.Fatalf("missing verified enrichment: %+v", record)
	}

	if record["request_id"] != response.Header().Get("X-Request-ID") ||
		strings.Contains(logs.String(), bobToken) {
		t.Fatal("bad correlation or credential logging")
	}
}

func TestRoutingErrorsHaveJSONEnvelopes(t *testing.T) {
	handler := testRouter(t, &operationSpy{}, false, io.Discard)

	for _, request := range []*http.Request{
		httptest.NewRequest("GET", "/missing", nil),
		httptest.NewRequest("DELETE", paths.ListPurchases.Path, nil),
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		var body api.ErrorResponse
		if err := json.Unmarshal(
			response.Body.Bytes(),
			&body,
		); err != nil ||
			body.Error.RequestID == "" {
			t.Fatalf("bad routing envelope: %s %v", response.Body, err)
		}

		if response.Code != 404 && response.Code != 405 {
			t.Fatalf("unexpected status: %d", response.Code)
		}
	}
}

func TestNegotiatedGzip(t *testing.T) {
	handler := testRouter(t, &operationSpy{}, false, io.Discard)
	request := httptest.NewRequest("GET", paths.ListPurchases.Path, nil)
	request.Header.Set("Authorization", "Bearer "+aliceToken)
	request.Header.Set("X-Organization-ID", "northstar")
	request.Header.Set("Accept-Encoding", "gzip")

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("gzip not negotiated: %+v", response.Header())
	}

	reader, err := gzip.NewReader(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	var body api.PurchasesResponse
	if err := json.NewDecoder(reader).
		Decode(&body); err != nil ||
		len(body.Data) != 1 {
		t.Fatalf("bad compressed JSON: %+v %v", body, err)
	}

	uncompressed := call(
		handler,
		paths.ListPurchases,
		aliceToken,
		"northstar",
		"",
	)
	if uncompressed.Header().Get("Content-Encoding") != "" {
		t.Fatal("compressed without negotiation")
	}
}
