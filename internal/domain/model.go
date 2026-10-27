// Package domain defines purchasing concepts without transport or persistence details.
package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

type Permission string

const (
	PurchaseCreate   Permission = "purchases:create"
	PurchaseRead     Permission = "purchases:read"
	PurchaseDecide   Permission = "purchases:decide"
	PurchaseExport   Permission = "purchases:export"
	MembershipManage Permission = "memberships:manage"
	AuditRead        Permission = "audit:read"
	AuditReadAll     Permission = "audit:read_all"
)

type Role string

const (
	Member        Role = "member"
	Approver      Role = "approver"
	Administrator Role = "administrator"
)

// Role grants are distinct from endpoint names and resource policies.
var rolePermissions = map[Role][]Permission{
	Member: {
		PurchaseCreate,
		PurchaseRead,
		AuditRead,
	},
	Approver: {
		PurchaseCreate,
		PurchaseRead,
		PurchaseDecide,
		PurchaseExport,
		AuditRead,
		AuditReadAll,
	},
	Administrator: {
		PurchaseCreate,
		PurchaseRead,
		PurchaseDecide,
		PurchaseExport,
		MembershipManage,
		AuditRead,
		AuditReadAll,
	},
}

func ValidRole(role Role) bool { _, ok := rolePermissions[role]; return ok }

type Principal struct {
	ActorID  string
	TenantID string
	Role     Role
}

func (p Principal) Can(permission Permission) bool {
	for _, grant := range rolePermissions[p.Role] {
		if grant == permission {
			return true
		}
	}

	return false
}

func (p Principal) Verified() bool { return p.ActorID != "" && p.TenantID != "" && ValidRole(p.Role) }

type Status string

const (
	Draft     Status = "draft"
	Submitted Status = "submitted"
	Approved  Status = "approved"
	Rejected  Status = "rejected"
)

type Purchase struct {
	ID          string
	TenantID    string
	CreatorID   string
	Title       string
	AmountCents int64
	Status      Status
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Membership struct {
	TenantID    string
	ActorID     string
	DisplayName string
	Role        Role
}

type AuditEvent struct {
	ID         string
	TenantID   string
	ActorID    string
	Action     string
	ResourceID string
	Outcome    string
	OccurredAt time.Time
}

var (
	ErrUnauthenticated = errors.New("authentication required")
	ErrForbidden       = errors.New("permission denied")
	ErrNotFound        = errors.New("resource not found")
	ErrConflict        = errors.New("resource changed or transition is invalid")
	ErrInvalid         = errors.New("invalid input")
	ErrFeatureDisabled = errors.New("feature disabled")
)

func ValidatePurchase(title string, cents int64) error {
	if strings.TrimSpace(title) == "" || len(title) > 160 ||
		!utf8.ValidString(title) ||
		strings.ContainsRune(title, 0) ||
		cents < 1 ||
		cents > 100_000_000 {
		return ErrInvalid
	}

	return nil
}
