// Package auth contains only an explicitly enabled local demo bearer authenticator.
package auth

import (
	"context"

	"github.com/plutoframework/pluto-example/internal/domain"
)

type MembershipReader interface {
	Membership(context.Context, string, string) (domain.Membership, error)
}

type Authenticator interface {
	Authenticate(context.Context, string, string) (domain.Principal, error)
}

type Demo struct {
	memberships MembershipReader
	tokens      map[string]string
}

func NewDemo(memberships MembershipReader) *Demo {
	return &Demo{
		memberships: memberships,
		tokens: map[string]string{
			"demo_alice_mJ7Kq9vL2xR8": "alice",
			"demo_bob_tW4nC6yP1sF3":   "bob",
			"demo_carol_hG8dA5uZ9eQ2": "carol",
			"demo_dana_bN6rV3kT7jS4":  "dana",
		},
	}
}

func (d *Demo) Authenticate(
	ctx context.Context,
	token, tenant string,
) (domain.Principal, error) {
	actor, ok := d.tokens[token]
	if !ok || tenant == "" {
		return domain.Principal{}, domain.ErrUnauthenticated
	}

	membership, err := d.memberships.Membership(ctx, tenant, actor)
	if err != nil {
		return domain.Principal{}, domain.ErrUnauthenticated
	}

	return domain.Principal{
		ActorID:  actor,
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
