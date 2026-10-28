package config

import "time"

type AuthConfig struct {
	DevelopmentSMTP string
	TrustedOrigins  []string
	SessionTTL      time.Duration
	IdleTimeout     time.Duration
}
