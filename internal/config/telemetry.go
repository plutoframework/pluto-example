package config

import "github.com/plutoframework/pluto-core/telemetry"

// LoadTelemetry reads only the application-prefixed telemetry settings once at
// startup. API and worker have separate service identities in the same environment.
func LoadTelemetry(service, environment string) (telemetry.Config, error) {
	return telemetry.FromEnv(Prefix, service, "development", environment)
}
