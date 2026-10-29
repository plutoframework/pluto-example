package config

import "testing"

func TestReviewProviderValidation(t *testing.T) {
	for _, c := range []ReviewConfig{{}, {Mode: "off"}, {Mode: "demo"}} {
		if err := c.Validate(); err != nil {
			t.Fatal(err)
		}
	}

	for _, c := range []ReviewConfig{
		{Mode: "unknown"},
		{Mode: "openai"},
		{Mode: "openai", APIKey: "synthetic", Model: "unsupported"},
	} {
		if c.Validate() == nil {
			t.Fatal("invalid provider configuration accepted")
		}
	}
}
