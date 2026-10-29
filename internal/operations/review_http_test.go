package operations

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/features"

	agent "github.com/plutoframework/pluto-agent"
	"github.com/plutoframework/pluto-agent/providers/openai"
)

func TestReviewWithMockedOpenAIHTTP(t *testing.T) {
	calls := 0

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++

			var request struct {
				Model     string `json:"model"`
				Store     bool   `json:"store"`
				MaxTokens int    `json:"max_completion_tokens"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Error(err)
			}

			if request.Model != "gpt-5.4" || request.Store ||
				request.MaxTokens != 512 {
				t.Error("incorrect bounded provider request")
			}

			w.Header().Set("Content-Type", "application/json")

			if calls == 1 {
				_, _ = w.Write(
					[]byte(
						`{"choices":[{"index":0,"finish_reason":"tool_calls","message":{"role":"assistant","tool_calls":[{"id":"read","type":"function","function":{"name":"read_purchase","arguments":"{}"}}]}}],"usage":{"prompt_tokens":10,"completion_tokens":10}}`,
					),
				)

				return
			}

			_, _ = w.Write(
				[]byte(
					`{"choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"{\"summary\":\"Check the justification.\",\"next_steps\":[\"Consult a human reviewer.\"],\"demo\":true}"}}],"usage":{"prompt_tokens":10,"completion_tokens":10}}`,
				),
			)
		}),
	)
	defer server.Close()

	repository := &purchaseFake{current: fixture(domain.Draft)}
	service := testService(t, repository)
	service.features = features.NewSnapshot(
		map[features.Key]bool{features.PurchaseReview: true},
	)
	service.ConfigureReview(func() (agent.Provider, error) {
		return openai.New(openai.Config{
			APIKey:                "synthetic",
			Model:                 "gpt-5.4",
			Endpoint:              server.URL,
			HTTPClient:            server.Client(),
			AllowInsecureLoopback: true,
		})
	}, false, nil)

	result, err := service.ReviewPurchase(
		context.Background(),
		principal("alice", domain.Member),
		"request-1",
	)
	if err != nil || calls != 2 || repository.writes != 0 || result.Demo {
		t.Fatalf(
			"HTTP review failed: %v calls=%d writes=%d demo=%v",
			err,
			calls,
			repository.writes,
			result.Demo,
		)
	}
}
