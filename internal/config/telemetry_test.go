package config

import "testing"

func TestTelemetryEnvironment(t *testing.T) {
	t.Setenv(Prefix+"OTEL_ENABLED", "false")
	t.Setenv(Prefix+"OTEL_SAMPLE_RATIO", "0.25")
	t.Setenv(Prefix+"OTEL_ENDPOINT", "http://127.0.0.1:4318")

	settings, err := LoadTelemetry("example-worker", "test")
	if err != nil {
		t.Fatal(err)
	}

	if settings.Enabled || settings.Service != "example-worker" ||
		settings.Environment != "test" ||
		settings.SampleRatio != 0.25 {
		t.Fatalf(
			"unexpected telemetry config: enabled=%v service=%s environment=%s ratio=%v",
			settings.Enabled,
			settings.Service,
			settings.Environment,
			settings.SampleRatio,
		)
	}

	t.Setenv(Prefix+"OTEL_ENABLED", "invalid")

	if _, err := LoadTelemetry("example-api", "test"); err == nil {
		t.Fatal("invalid telemetry setting accepted")
	}
}
