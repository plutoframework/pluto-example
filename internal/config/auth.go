package config

import "time"

type AuthConfig struct {
	TrustedOrigins []string
	SessionTTL     time.Duration
	IdleTimeout    time.Duration
}
