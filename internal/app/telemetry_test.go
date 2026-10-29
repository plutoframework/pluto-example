package app

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/plutoframework/pluto-example/internal/config"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/features"

	"github.com/plutoframework/pluto-core/telemetry"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type retainingExporter struct{ *tracetest.InMemoryExporter }

func (retainingExporter) Shutdown(context.Context) error { return nil }

func TestTelemetryRoutesAndLogCorrelation(t *testing.T) {
	exporter := retainingExporter{tracetest.NewInMemoryExporter()}

	runtime, err := telemetry.New(context.Background(), telemetry.Config{
		Enabled: true, Service: "example-test", SampleRatio: 1,
		TraceExporter: exporter, MetricReader: sdkmetric.NewManualReader(),
	})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = runtime.Shutdown(context.Background()) }()

	policy, err := domain.NewPolicy()
	if err != nil {
		t.Fatal(err)
	}

	var logs bytes.Buffer

	logger := slog.New(runtime.Handler(slog.NewJSONHandler(&logs, nil)))
	router := RoutesWithTelemetry(
		config.ServerConfig{MaxBodyBytes: 512},
		logger,
		SessionAuthentication{},
		features.NewSnapshot(nil),
		&operationSpy{},
		policy,
		runtime,
	)
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/purchases/private-id?token=private-token",
		nil,
	)
	request.Header.Set(
		"traceparent",
		"00-0123456789abcdef0123456789abcdef-0123456789abcdef-01",
	)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", response.Code)
	}

	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 || spans[0].Name != "GET /api/purchases/{purchaseID}" {
		t.Fatalf("unexpected server spans: %+v", spans)
	}

	for _, attr := range spans[0].Attributes {
		value := attr.Value.AsString()
		if strings.Contains(value, "private-id") ||
			strings.Contains(value, "private-token") {
			t.Fatal("raw path/query in span")
		}
	}

	if !strings.Contains(
		logs.String(),
		`"trace_id":"0123456789abcdef0123456789abcdef"`,
	) {
		t.Fatalf("missing correlated trace: %s", logs.String())
	}

	if strings.Contains(logs.String(), "private-id") ||
		strings.Contains(logs.String(), "private-token") {
		t.Fatal("request data in diagnostic log")
	}
}
