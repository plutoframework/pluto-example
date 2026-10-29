package operations

import (
	"context"
	"errors"
	"testing"

	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/features"

	agent "github.com/plutoframework/pluto-agent"
	"github.com/plutoframework/pluto-agent/providers/scripted"
)

func TestReviewAuthorizationBeforeProvider(t *testing.T) {
	for _, scenario := range []struct {
		name    string
		actor   domain.Principal
		enabled bool
		id      string
	}{
		{"anonymous", domain.Principal{}, true, "request-1"},
		{"other owner", principal("mallory", domain.Member), true, "request-1"},
		{"missing", principal("alice", domain.Member), true, "missing"},
		{"disabled", principal("alice", domain.Member), false, "request-1"},
		{"other tenant", domain.Principal{ActorID: "alice", TenantID: "other", Role: domain.Administrator}, true, "request-1"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			repository := &purchaseFake{current: fixture(domain.Draft)}
			service := testService(t, repository)
			service.features = features.NewSnapshot(map[features.Key]bool{
				features.PurchaseReview: scenario.enabled,
			})
			calls := 0

			service.ConfigureReview(func() (agent.Provider, error) {
				calls++
				return DemoReviewProvider()
			}, true, nil)

			_, err := service.ReviewPurchase(
				context.Background(),
				scenario.actor,
				scenario.id,
			)
			if err == nil || calls != 0 || repository.writes != 0 {
				t.Fatalf(
					"authorization leaked: err=%v calls=%d writes=%d",
					err,
					calls,
					repository.writes,
				)
			}
		})
	}
}

func TestReviewReadOnlyAndRunNarrowing(t *testing.T) {
	repository := &purchaseFake{current: fixture(domain.Draft)}
	service := testService(t, repository)
	service.features = features.NewSnapshot(map[features.Key]bool{
		features.PurchaseReview: true,
	})
	service.ConfigureReview(DemoReviewProvider, true, nil)

	result, err := service.ReviewPurchase(
		context.Background(),
		principal("alice", domain.Member),
		"request-1",
	)
	if err != nil || !result.Demo || result.Summary == "" ||
		repository.writes != 0 {
		t.Fatalf(
			"demo review: %+v %v writes=%d",
			result,
			err,
			repository.writes,
		)
	}

	service.ConfigureReview(func() (agent.Provider, error) {
		return scripted.New(scripted.Calls(agent.ToolCall{
			ID: "forbidden", Name: "pay_purchase", Arguments: `{}`,
		})), nil
	}, false, nil)

	_, err = service.ReviewPurchase(
		context.Background(),
		principal("admin", domain.Administrator),
		"request-1",
	)
	if !errors.Is(err, agent.ErrDenied) || repository.writes != 0 {
		t.Fatalf("privileged run escaped: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = service.ReviewPurchase(
		ctx,
		principal("admin", domain.Administrator),
		"request-1",
	)
	if err == nil || repository.writes != 0 {
		t.Fatalf("canceled run succeeded: %v", err)
	}
}

func TestReviewRejectsInvalidOutputAndExcessTools(t *testing.T) {
	for _, scenario := range []struct {
		name    string
		factory func() (agent.Provider, error)
	}{
		{name: "invalid output", factory: func() (agent.Provider, error) {
			return scripted.New(scripted.Final(`{"summary":"","next_steps":[],"demo":false}`)), nil
		}},
		{name: "excess tool calls", factory: func() (agent.Provider, error) {
			return scripted.New(scripted.Calls(
				agent.ToolCall{ID: "one", Name: "read_purchase", Arguments: `{}`},
				agent.ToolCall{ID: "two", Name: "read_purchase", Arguments: `{}`},
			)), nil
		}},
		{name: "provider error", factory: func() (agent.Provider, error) {
			return nil, errors.New("synthetic provider failure")
		}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			repository := &purchaseFake{current: fixture(domain.Draft)}
			service := testService(t, repository)
			service.features = features.NewSnapshot(map[features.Key]bool{
				features.PurchaseReview: true,
			})
			service.ConfigureReview(scenario.factory, false, nil)

			_, err := service.ReviewPurchase(context.Background(),
				principal("alice", domain.Member), "request-1")
			if err == nil || repository.writes != 0 {
				t.Fatalf(
					"invalid run succeeded: %v writes=%d",
					err,
					repository.writes,
				)
			}
		})
	}
}
