package operations

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/plutoframework/pluto-example/internal/domain"
)

func TestAccessUsesCompiledPolicy(t *testing.T) {
	policy, _ := domain.NewPolicy()
	service := AccessService{Policy: policy}
	principal := domain.Principal{
		ActorID:  "alice",
		TenantID: "northstar",
		Role:     domain.Member,
	}

	value, err := service.Access(context.Background(), principal)
	if err != nil || value.OrganizationID != "northstar" ||
		value.ActorID != "alice" ||
		!slices.Contains(value.Permissions, domain.PurchaseCreate) ||
		slices.Contains(value.Permissions, domain.MembershipManage) ||
		len(value.Roles) != 0 {
		t.Fatal(value, err)
	}

	principal.Role = domain.Administrator

	value, err = service.Access(context.Background(), principal)
	if err != nil ||
		!slices.Contains(value.Permissions, domain.PurchaseCreate) ||
		!slices.Contains(value.Permissions, domain.MembershipManage) ||
		len(value.Roles) != 3 {
		t.Fatal(value, err)
	}

	principal.Role = "unknown"
	if _, err := service.Access(
		context.Background(),
		principal,
	); !errors.Is(
		err,
		domain.ErrForbidden,
	) {
		t.Fatal("unknown role granted visibility", err)
	}

	principal.ActorID = ""
	if _, err := service.Access(
		context.Background(),
		principal,
	); !errors.Is(
		err,
		domain.ErrUnauthenticated,
	) {
		t.Fatal("unverified scope exposed", err)
	}
}
