// Package auth adapts verified sessions to application membership.
package auth

import (
	"context"

	"github.com/plutoframework/pluto-example/internal/domain"

	"github.com/plutoframework/pluto-auth/httpauth"
)

type MembershipReader interface {
	Membership(context.Context, string, string) (domain.Membership, error)
}

type Authenticator interface {
	Authenticate(context.Context, string, string) (domain.Principal, error)
}

// Sessions resolves the exact membership only after HTTP session verification.
// Request tenant selection never establishes a membership by itself.
type Sessions struct{ Memberships MembershipReader }

func (s Sessions) Authenticate(
	ctx context.Context,
	_ string,
	tenant string,
) (domain.Principal, error) {
	session, ok := httpauth.FromContext(ctx)
	if !ok || tenant == "" {
		return domain.Principal{}, domain.ErrUnauthenticated
	}

	membership, err := s.Memberships.Membership(ctx, tenant, session.SubjectID)
	if err != nil {
		return domain.Principal{}, domain.ErrUnauthenticated
	}

	if membership.ActorID != session.SubjectID ||
		membership.TenantID != tenant {
		return domain.Principal{}, domain.ErrUnauthenticated
	}

	return domain.Principal{
		ActorID:  session.SubjectID,
		TenantID: membership.TenantID,
		Role:     membership.Role,
	}, nil
}

func DemoMemberships() []domain.Membership {
	return []domain.Membership{
		{
			TenantID:    "northstar",
			ActorID:     "alice",
			DisplayName: "Alice Chen",
			Role:        domain.Member,
		},
		{
			TenantID:    "northstar",
			ActorID:     "bob",
			DisplayName: "Bob Rivera",
			Role:        domain.Approver,
		},
		{
			TenantID:    "northstar",
			ActorID:     "carol",
			DisplayName: "Carol Okafor",
			Role:        domain.Administrator,
		},
		{
			TenantID:    "harbor",
			ActorID:     "dana",
			DisplayName: "Dana Patel",
			Role:        domain.Administrator,
		},
		{
			TenantID:    "harbor",
			ActorID:     "alice",
			DisplayName: "Alice Chen",
			Role:        domain.Member,
		},
	}
}
