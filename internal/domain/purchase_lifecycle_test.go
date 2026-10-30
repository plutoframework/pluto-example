package domain_test

import (
	"errors"
	"testing"

	"github.com/plutoframework/pluto-example/internal/domain"

	"github.com/plutoframework/pluto-core/state"
)

func TestPurchaseLifecycleGuards(t *testing.T) {
	machine, err := domain.NewPurchaseLifecycle()
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name  string
		from  domain.Status
		event domain.PurchaseEvent
		facts domain.PurchaseFacts
		want  error
	}{
		{
			"missing actor", domain.Draft, domain.SubmitPurchase,
			domain.PurchaseFacts{CreatorID: "alice", AmountCents: 1},
			domain.ErrInvalid,
		},
		{
			"invalid amount", domain.Submitted, domain.ApprovePurchase,
			domain.PurchaseFacts{ActorID: "bob", CreatorID: "alice"},
			domain.ErrInvalid,
		},
		{
			"foreign submission", domain.Draft, domain.SubmitPurchase,
			domain.PurchaseFacts{ActorID: "bob", CreatorID: "alice", AmountCents: 1},
			domain.ErrNotFound,
		},
		{
			"self approval", domain.Submitted, domain.ApprovePurchase,
			domain.PurchaseFacts{ActorID: "alice", CreatorID: "alice", AmountCents: 1},
			domain.ErrForbidden,
		},
		{
			"unknown state", "corrupt", domain.SubmitPurchase,
			domain.PurchaseFacts{},
			state.ErrState,
		},
		{
			"terminal", domain.Rejected, domain.ApprovePurchase,
			domain.PurchaseFacts{},
			state.ErrTransition,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			next, err := machine.Next(test.from, test.event, test.facts)
			if next != test.from || !errors.Is(err, test.want) {
				t.Fatalf("next=%s error=%v", next, err)
			}
		})
	}
}
