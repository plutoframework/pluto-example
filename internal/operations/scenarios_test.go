package operations_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/plutoframework/pluto-example/internal/auth"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/features"
	"github.com/plutoframework/pluto-example/internal/operations"
	"github.com/plutoframework/pluto-example/internal/storage/memory"
)

func TestOrganizationWorkflow(t *testing.T) {
	ctx := context.Background()
	store := memory.New(auth.DemoMemberships())
	service := newTestService(t,
		store,
		features.NewSnapshot(
			map[features.Key]bool{features.PurchaseExport: true},
		),
	)
	alice := domain.Principal{
		ActorID:  "alice",
		TenantID: "northstar",
		Role:     domain.Member,
	}
	bob := domain.Principal{
		ActorID:  "bob",
		TenantID: "northstar",
		Role:     domain.Approver,
	}
	dana := domain.Principal{
		ActorID:  "dana",
		TenantID: "harbor",
		Role:     domain.Administrator,
	}

	request, err := service.Create(ctx, alice, "Field workstation", 185000)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.Submit(ctx, alice, request.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := service.Decide(
		ctx,
		bob,
		request.ID,
		domain.Approved,
	); err != nil {
		t.Fatal(err)
	}

	for _, actor := range []domain.Principal{
		dana,
		{
			ActorID:  "alice",
			TenantID: "harbor",
			Role:     domain.Member,
		},
	} {
		if _, err := service.Get(
			ctx,
			actor,
			request.ID,
		); !errors.Is(
			err,
			domain.ErrNotFound,
		) {
			t.Fatalf("tenant escape: %v", err)
		}

		listed, err := service.List(ctx, actor)
		if err != nil || len(listed) != 0 {
			t.Fatalf("tenant list leak: %+v %v", listed, err)
		}

		audit, err := service.Audit(ctx, actor)
		if err != nil || len(audit) != 0 {
			t.Fatalf("tenant audit leak: %+v %v", audit, err)
		}
	}

	own, err := service.Audit(ctx, alice)
	if err != nil || len(own) != 2 {
		t.Fatalf("member must see only their two actions: %+v %v", own, err)
	}

	all, err := service.Audit(ctx, bob)
	if err != nil || len(all) != 3 {
		t.Fatalf("approver must see three org events: %+v %v", all, err)
	}

	export, err := service.Export(ctx, bob)
	if err != nil || len(export) != 1 || export[0].Status != domain.Approved {
		t.Fatalf("bad export: %+v %v", export, err)
	}
}

func TestMembershipPolicies(t *testing.T) {
	ctx := context.Background()
	store := memory.New(auth.DemoMemberships())
	service := newTestService(t, store, features.NewSnapshot(nil))
	carol := domain.Principal{
		ActorID:  "carol",
		TenantID: "northstar",
		Role:     domain.Administrator,
	}

	alice := domain.Principal{
		ActorID:  "alice",
		TenantID: "northstar",
		Role:     domain.Member,
	}
	if err := service.ChangeRole(
		ctx,
		alice,
		"bob",
		domain.Member,
	); !errors.Is(
		err,
		domain.ErrForbidden,
	) {
		t.Fatalf("member changed role: %v", err)
	}

	if err := service.ChangeRole(
		ctx,
		carol,
		"carol",
		domain.Member,
	); !errors.Is(
		err,
		domain.ErrForbidden,
	) {
		t.Fatalf("self role change: %v", err)
	}

	if err := service.ChangeRole(
		ctx,
		carol,
		"dana",
		domain.Member,
	); !errors.Is(
		err,
		domain.ErrNotFound,
	) {
		t.Fatalf("cross tenant role change: %v", err)
	}

	if err := service.ChangeRole(
		ctx,
		carol,
		"alice",
		"owner",
	); !errors.Is(
		err,
		domain.ErrInvalid,
	) {
		t.Fatalf("invalid role: %v", err)
	}

	if err := service.ChangeRole(
		ctx,
		carol,
		"alice",
		domain.Approver,
	); err != nil {
		t.Fatal(err)
	}

	membership, _ := store.Membership(ctx, "northstar", "alice")
	if membership.Role != domain.Approver {
		t.Fatal("role not saved")
	}

	events, _ := service.Audit(ctx, carol)
	if len(events) != 1 || events[0].Action != "membership.role_changed" ||
		events[0].ActorID != "carol" {
		t.Fatalf("bad role audit: %+v", events)
	}
}

func TestOptimisticWriteHasOneWinnerAndOneAudit(t *testing.T) {
	ctx := context.Background()
	store := memory.New(auth.DemoMemberships())
	service := newTestService(t, store, features.NewSnapshot(nil))
	actor := domain.Principal{
		ActorID:  "alice",
		TenantID: "northstar",
		Role:     domain.Member,
	}

	purchase, err := service.Create(ctx, actor, "Inspection kit", 5000)
	if err != nil {
		t.Fatal(err)
	}

	purchase.Status = domain.Submitted
	purchase.Version = 2
	results := make(chan error, 2)

	var wait sync.WaitGroup
	for i := 0; i < 2; i++ {
		wait.Add(1)

		go func() {
			defer wait.Done()

			results <- store.SavePurchase(ctx, purchase, 1, domain.AuditEvent{
				TenantID: "northstar",
				ActorID:  "alice",
				Action:   "purchase.submitted",
			})
		}()
	}

	wait.Wait()
	close(results)

	successes, conflicts := 0, 0

	for err := range results {
		if err == nil {
			successes++
		} else if errors.Is(err, domain.ErrConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}

	if successes != 1 || conflicts != 1 {
		t.Fatalf("winners=%d conflicts=%d", successes, conflicts)
	}

	events, _ := store.Audit(ctx, "northstar", "")
	if len(events) != 2 {
		t.Fatalf("stale write appended an audit: %+v", events)
	}
}

func newTestService(
	t testing.TB,
	repository operations.Repository,
	evaluator features.Evaluator,
) *operations.Service {
	t.Helper()

	policy, err := domain.NewPolicy()
	if err != nil {
		t.Fatal(err)
	}

	return operations.New(repository, evaluator, policy)
}
