package operations

import (
	"context"

	"github.com/plutoframework/pluto-example/internal/domain"
)

func (s *Service) Memberships(
	ctx context.Context,
	principal domain.Principal,
) ([]domain.Membership, error) {
	if err := require(principal, domain.MembershipManage); err != nil {
		return nil, err
	}

	return s.repository.Memberships(ctx, principal.TenantID)
}

func (s *Service) ChangeRole(
	ctx context.Context,
	principal domain.Principal,
	actorID string,
	role domain.Role,
) error {
	if err := require(principal, domain.MembershipManage); err != nil {
		return err
	}

	if !domain.ValidRole(role) || actorID == "" {
		return domain.ErrInvalid
	}

	// Administrators cannot change their own role. This is a resource policy,
	// not a guarantee against concurrent administrators demoting one another.
	if actorID == principal.ActorID {
		return domain.ErrForbidden
	}

	return s.repository.ChangeRole(
		ctx,
		principal.TenantID,
		actorID,
		role,
		s.event(principal, "membership.role_changed", actorID),
	)
}

func (s *Service) Audit(
	ctx context.Context,
	principal domain.Principal,
) ([]domain.AuditEvent, error) {
	if err := require(principal, domain.AuditRead); err != nil {
		return nil, err
	}

	actorFilter := principal.ActorID
	if principal.Can(domain.AuditReadAll) {
		actorFilter = ""
	}

	return s.repository.Audit(ctx, principal.TenantID, actorFilter)
}
