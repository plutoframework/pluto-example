// Package features is the adapter boundary for future remote evaluation providers.
package features

import (
	"github.com/plutoframework/pluto-example/internal/domain"

	corefeatures "github.com/plutoframework/pluto-core/features"
)

type Key = corefeatures.Key

const PurchaseExport Key = "purchase_export"

const PurchaseReview Key = "purchase_review"

type Definition = corefeatures.Definition

var Export = Definition{
	Key:           PurchaseExport,
	Name:          "Purchase request export",
	Description:   "Allow authorized approvers to export the active organization's requests.",
	Default:       false,
	ClientVisible: true,
}

var internalDiagnostics = Definition{
	Key:           "internal_diagnostics",
	Name:          "Internal diagnostics",
	Description:   "An internal-only provider example; never exposed through the client endpoint.",
	Default:       true,
	ClientVisible: false,
}

func Definitions() []Definition {
	return []Definition{
		Export,
		{
			Key: PurchaseReview, Name: "Purchase review",
			Description: "Read-only purchase review suggestions.",
			Default:     false, ClientVisible: true,
		},
		internalDiagnostics,
	}
}

type Evaluator interface {
	Enabled(Key, domain.Principal) bool
}

// Snapshot adapts Core evaluation to the application's verified principal.
type Snapshot struct{ snapshot *corefeatures.Snapshot }

func NewSnapshot(overrides map[Key]bool) *Snapshot {
	return &Snapshot{
		snapshot: corefeatures.MustSnapshot(Definitions(), overrides),
	}
}

func (s *Snapshot) Enabled(key Key, principal domain.Principal) bool {
	return s != nil && principal.Verified() && s.snapshot.Enabled(key)
}

func Visible(evaluator Evaluator, principal domain.Principal) map[string]bool {
	return corefeatures.Visible(
		Definitions(),
		corefeatures.EvaluateFunc(func(key Key) bool {
			return evaluator != nil && evaluator.Enabled(key, principal)
		}),
	)
}
