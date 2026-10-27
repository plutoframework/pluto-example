package domain

import rbac "github.com/plutoframework/pluto-rbac"

// PermissionDefinition supports human-readable role assignment interfaces.
// IDs are authorization contracts; names may evolve without changing policies.
type PermissionDefinition = rbac.Permission

func Permissions() []PermissionDefinition {
	return []PermissionDefinition{
		{
			ID:          PurchaseCreate,
			Name:        "Create and submit requests",
			Description: "Create drafts and submit requests that you own.",
		},
		{
			ID:          PurchaseRead,
			Name:        "Read requests",
			Description: "Read requests allowed by ownership and organization policies.",
		},
		{
			ID:          PurchaseDecide,
			Name:        "Review organization requests",
			Description: "View organization requests and decide submitted requests made by others.",
		},
		{
			ID:          PurchaseExport,
			Name:        "Export organization requests",
			Description: "Export visible organization requests when the capability is enabled.",
		},
		{
			ID:          MembershipManage,
			Name:        "Manage organization roles",
			Description: "List memberships and change another existing member's role.",
		},
		{
			ID:          AuditRead,
			Name:        "Read personal audit history",
			Description: "Read your own actions within the active organization.",
		},
		{
			ID:          AuditReadAll,
			Name:        "Read organization audit history",
			Description: "Read every actor's audit events within the active organization.",
		},
	}
}

type RoleDefinition = rbac.Role

func Roles() []RoleDefinition {
	return []RoleDefinition{
		{
			ID:          Member,
			Name:        "Member",
			Description: "Request purchases and follow your own activity.",
			Permissions: []Permission{PurchaseCreate, PurchaseRead, AuditRead},
		},
		{
			ID:          Approver,
			Inherits:    []Role{Member},
			Name:        "Approver",
			Description: "Review requests from colleagues and inspect organization history.",
			Permissions: []Permission{
				PurchaseDecide,
				PurchaseExport,
				AuditReadAll,
			},
		},
		{
			ID:          Administrator,
			Inherits:    []Role{Approver},
			Name:        "Administrator",
			Description: "Manage organization roles in addition to reviewing purchases.",
			Permissions: []Permission{MembershipManage},
		},
	}
}

// NewPolicy validates the application catalog once for constructor injection.
func NewPolicy() (*rbac.Policy, error) {
	return rbac.New(rbac.Definition{
		Permissions: Permissions(),
		Roles:       Roles(),
	})
}
