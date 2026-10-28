// Package config loads runtime settings once. pluto.toml is development metadata.
package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	jobs "github.com/plutoframework/pluto-jobs"
)

const Prefix = "PLUTO_EXAMPLE_"

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Logging  LoggingConfig
	Features FeaturesConfig
	DemoAuth bool
	Auth     AuthConfig
	Jobs     jobs.Config
}

func Load() (Config, error) { return LoadFrom(os.LookupEnv) }

type Lookup func(string) (string, bool)

func LoadFrom(lookup Lookup) (Config, error) {
	config := Config{
		Jobs: jobs.DefaultConfig(),
		Server: ServerConfig{
			Address:      "127.0.0.1:8080",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
			MaxBodyBytes: 16 * 1024,
			Gzip:         true,
		},
		Database: DatabaseConfig{Mode: "postgres"},
		Auth: AuthConfig{
			DevelopmentSMTP: "127.0.0.1:1025",
			TrustedOrigins:  []string{"http://127.0.0.1:8080"},
			SessionTTL:      24 * time.Hour,
			IdleTimeout:     30 * time.Minute,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Stdout: true,
		},
	}
	text := func(name string, target *string) {
		if value, ok := lookup(Prefix + name); ok {
			*target = value
		}
	}
	text("SERVER_ADDRESS", &config.Server.Address)
	text("DATABASE_MODE", &config.Database.Mode)
	text("DATABASE_URL", &config.Database.URL)
	text("AUTH_DEVELOPMENT_SMTP", &config.Auth.DevelopmentSMTP)
	text("LOG_LEVEL", &config.Logging.Level)
	text("LOG_FILE", &config.Logging.File)

	if value, ok := lookup(Prefix + "AUTH_TRUSTED_ORIGINS"); ok {
		config.Auth.TrustedOrigins = strings.Split(value, ",")
	}

	for _, setting := range []struct {
		name   string
		target *time.Duration
	}{
		{"AUTH_SESSION_TTL", &config.Auth.SessionTTL},
		{"AUTH_IDLE_TIMEOUT", &config.Auth.IdleTimeout},
	} {
		if value, ok := lookup(Prefix + setting.name); ok {
			parsed, err := time.ParseDuration(value)
			if err != nil {
				return Config{}, fmt.Errorf(
					"%s%s must be a duration",
					Prefix,
					setting.name,
				)
			}

			*setting.target = parsed
		}
	}

	for _, setting := range []struct {
		name   string
		target *bool
	}{
		{
			"DEMO_AUTH",
			&config.DemoAuth,
		},
		{
			"SERVER_GZIP",
			&config.Server.Gzip,
		},
		{
			"LOG_STDOUT",
			&config.Logging.Stdout,
		},
	} {
		if value, ok := lookup(Prefix + setting.name); ok {
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return Config{}, fmt.Errorf(
					"%s%s must be boolean",
					Prefix,
					setting.name,
				)
			}

			*setting.target = parsed
		}
	}

	if value, ok := lookup(Prefix + "FEATURE_PURCHASE_EXPORT"); ok {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return Config{}, fmt.Errorf(
				"%sFEATURE_PURCHASE_EXPORT must be boolean",
				Prefix,
			)
		}

		config.Features.PurchaseExport = &parsed
	}

	if err := config.Validate(); err != nil {
		return Config{}, err
	}

	return config, nil
}

func (c Config) Validate() error {
	if c.Auth.SessionTTL <= 0 || c.Auth.IdleTimeout < 0 ||
		c.Auth.IdleTimeout > c.Auth.SessionTTL ||
		len(c.Auth.TrustedOrigins) == 0 {
		return fmt.Errorf("invalid AUTH session configuration")
	}

	host, _, err := net.SplitHostPort(c.Server.Address)
	if err != nil {
		return fmt.Errorf("SERVER_ADDRESS must be host:port")
	}

	// This prototype has only local demo authentication. Bind only to loopback
	// while it is enabled, including when using the PostgreSQL adapter.
	address := net.ParseIP(host)

	if !c.DemoAuth {
		return fmt.Errorf(
			"set %sDEMO_AUTH=true for this local-only prototype",
			Prefix,
		)
	}

	if address == nil || !address.IsLoopback() {
		return fmt.Errorf(
			"demo authentication requires a numeric loopback SERVER_ADDRESS",
		)
	}

	if c.Database.Mode != "postgres" {
		return fmt.Errorf("local credentials require DATABASE_MODE=postgres")
	}

	if c.Database.Mode == "postgres" &&
		strings.TrimSpace(c.Database.URL) == "" {
		return fmt.Errorf("DATABASE_URL is required for postgres")
	}

	switch c.Logging.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("LOG_LEVEL must be debug, info, warn or error")
	}

	if !c.Logging.Stdout && c.Logging.File == "" {
		return fmt.Errorf("at least one log destination is required")
	}

	return nil
}
