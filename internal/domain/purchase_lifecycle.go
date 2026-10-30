package domain

import "github.com/plutoframework/pluto-core/state"

type PurchaseEvent string

const (
	SubmitPurchase  PurchaseEvent = "submit"
	ApprovePurchase PurchaseEvent = "approve"
	RejectPurchase  PurchaseEvent = "reject"
)

// PurchaseFacts is a snapshot of verified identity and the loaded purchase.
// Permissions and tenant visibility are checked by operations before evaluation.
type PurchaseFacts struct {
	ActorID     string
	CreatorID   string
	AmountCents int64
}

func NewPurchaseLifecycle() (*state.Machine[Status, PurchaseEvent, PurchaseFacts], error) {
	return state.New(state.Definition[Status, PurchaseEvent, PurchaseFacts]{
		States: []Status{Draft, Submitted, Approved, Rejected},
		Transitions: []state.Transition[Status, PurchaseEvent, PurchaseFacts]{
			{
				From:  Draft,
				On:    SubmitPurchase,
				To:    Submitted,
				Guard: canSubmitPurchase,
			},
			{
				From:  Submitted,
				On:    ApprovePurchase,
				To:    Approved,
				Guard: canDecidePurchase,
			},
			{
				From:  Submitted,
				On:    RejectPurchase,
				To:    Rejected,
				Guard: canDecidePurchase,
			},
		},
	})
}

func validPurchaseFacts(facts PurchaseFacts) bool {
	return facts.ActorID != "" && facts.CreatorID != "" && facts.AmountCents > 0
}

func canSubmitPurchase(facts PurchaseFacts) error {
	if !validPurchaseFacts(facts) {
		return ErrInvalid
	}

	if facts.ActorID != facts.CreatorID {
		return ErrNotFound
	}

	return nil
}

func canDecidePurchase(facts PurchaseFacts) error {
	if !validPurchaseFacts(facts) {
		return ErrInvalid
	}

	if facts.ActorID == facts.CreatorID {
		return ErrForbidden
	}

	return nil
}
