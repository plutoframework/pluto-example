package operations

import (
	"context"

	"github.com/plutoframework/pluto-example/internal/domain"

	"github.com/plutoframework/pluto-core/pagination"
)

// Repository describes only the purchasing capabilities consumed by this service.
// Every read and write is scoped to a verified tenant. Implementations must keep a
// successful purchase write and its event atomic and reject stale versions.
type Repository interface {
	PurchasePage(
		context.Context,
		string,
		string,
		pagination.Request,
	) ([]domain.Purchase, error)
	MembershipPage(
		context.Context,
		string,
		pagination.Request,
	) ([]domain.Membership, error)
	AuditPage(
		context.Context,
		string,
		string,
		pagination.Request,
	) ([]domain.AuditEvent, error)
	Membership(context.Context, string, string) (domain.Membership, error)
	Memberships(context.Context, string) ([]domain.Membership, error)
	ChangeRole(
		context.Context,
		string,
		string,
		domain.Role,
		domain.AuditEvent,
	) error
	Purchase(context.Context, string, string) (domain.Purchase, error)
	Purchases(context.Context, string) ([]domain.Purchase, error)
	SavePurchase(
		context.Context,
		domain.Purchase,
		int64,
		domain.AuditEvent,
	) error
	Audit(context.Context, string, string) ([]domain.AuditEvent, error)
}
