package config

import (
	"fmt"

	"github.com/plutoframework/pluto-agent/providers/openai"
)

// ReviewConfig is runtime-only. Credentials never belong in pluto.toml.
type ReviewConfig struct {
	Mode   string
	APIKey string
	Model  string
}

func (c ReviewConfig) Validate() error {
	switch c.Mode {
	case "", "off", "demo":
		return nil
	case "openai":
		_, err := openai.New(openai.Config{APIKey: c.APIKey, Model: c.Model})
		if err != nil {
			return fmt.Errorf("invalid AGENT_OPENAI_API_KEY or AGENT_MODEL")
		}

		return nil
	default:
		return fmt.Errorf("AGENT_MODE must be off, demo or openai")
	}
}
