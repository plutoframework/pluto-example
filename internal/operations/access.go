package operations

import (
	"context"

	"github.com/plutoframework/pluto-example/internal/domain"

	rbac "github.com/plutoframework/pluto-rbac"
)

type AccessService struct {
	Policy *rbac.Policy
}

// Access describes this request's verified membership. UI visibility is advisory;
// every business operation still checks its own current authorization and policy.
func (s AccessService) Access(
	ctx context.Context,
	principal domain.Principal,
) (domain.Access, error) {
	if err := ctx.Err(); err != nil {
		return domain.Access{}, err
	}

	if !principal.Verified() {
		return domain.Access{}, domain.ErrUnauthenticated
	}

	if s.Policy == nil || !domain.ValidRole(s.Policy, principal.Role) {
		return domain.Access{}, domain.ErrForbidden
	}

	result := domain.Access{
		ActorID:        principal.ActorID,
		OrganizationID: principal.TenantID,
		Role:           principal.Role,
		Permissions:    []domain.Permission{},
		Roles:          []domain.RoleDefinition{},
	}
	for _, permission := range domain.Permissions() {
		if principal.Can(s.Policy, permission.ID) {
			result.Permissions = append(result.Permissions, permission.ID)
		}
	}

	for _, role := range domain.Roles() {
		if role.ID == principal.Role {
			result.RoleName = role.Name
		}

		if principal.Can(s.Policy, domain.MembershipManage) {
			result.Roles = append(result.Roles, role)
		}
	}

	return result, nil
}
