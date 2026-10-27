// Package features is the adapter boundary for future remote evaluation providers.
package features

import "github.com/plutoframework/pluto-example/internal/domain"

type Key string

const PurchaseExport Key = "purchase_export"

type Definition struct {
	Key           Key
	Name          string
	Description   string
	Default       bool
	ClientVisible bool
}

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
		internalDiagnostics,
	}
}

type Evaluator interface {
	Enabled(Key, domain.Principal) bool
}

// Snapshot copies only declared startup values; absent values use each definition's fallback.
// An explicit false is preserved. This is a local provider, not OpenFeature or OFREP.
type Snapshot struct{ values map[Key]bool }

func NewSnapshot(overrides map[Key]bool) *Snapshot {
	values := make(map[Key]bool)

	for _, definition := range Definitions() {
		value, present := overrides[definition.Key]
		if !present {
			value = definition.Default
		}

		values[definition.Key] = value
	}

	return &Snapshot{values: values}
}

func (s *Snapshot) Enabled(key Key, principal domain.Principal) bool {
	return principal.Verified() && s.values[key]
}

func Visible(evaluator Evaluator, principal domain.Principal) map[string]bool {
	values := make(map[string]bool)

	for _, definition := range Definitions() {
		if definition.ClientVisible {
			values[string(definition.Key)] = evaluator.Enabled(
				definition.Key,
				principal,
			)
		}
	}

	return values
}
