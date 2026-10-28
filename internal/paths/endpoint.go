// Package paths is the sole route catalog. It never imports handlers.
package paths

import (
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/features"
)

type Endpoint struct {
	Collection    bool
	Idempotent    bool
	ID            string
	Name          string
	Description   string
	Method        string
	Path          string
	Authenticated bool
	SessionOnly   bool
	Permission    domain.Permission
	Feature       features.Key
	Request       any
	Response      any
	SuccessStatus int
}

func All() []Endpoint {
	return []Endpoint{
		GetOrganization,
		Enroll,
		RequestVerification,
		VerifyEmail,
		SignIn,
		RequestRecovery,
		ResetPassword,
		Logout,
		GetSession,
		DeleteSession,
		CreatePurchase,
		ListPurchases,
		GetPurchase,
		SubmitPurchase,
		DecidePurchase,
		ExportPurchases,
		ListMemberships,
		ChangeRole,
		GetFeatures,
		ListAudit,
	}
}
