package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/features"
	"github.com/plutoframework/pluto-example/internal/operations"
)

// Both operations must evaluate the SAME snapshot before either can commit.
// This proves a successful state decision cannot bypass the storage version fence.
type lifecycleBarrier struct {
	*Store
	loaded  chan struct{}
	release chan struct{}
}

func (b *lifecycleBarrier) Purchase(
	ctx context.Context,
	tenant, id string,
) (domain.Purchase, error) {
	purchase, err := b.Store.Purchase(ctx, tenant, id)
	if err != nil {
		return purchase, err
	}

	select {
	case b.loaded <- struct{}{}:
	case <-ctx.Done():
		return domain.Purchase{}, ctx.Err()
	}

	select {
	case <-b.release:
		return purchase, nil
	case <-ctx.Done():
		return domain.Purchase{}, ctx.Err()
	}
}

func TestPostgresLifecycleConcurrentDecisions(t *testing.T) {
	store := isolatedStore(t)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	purchase := createJobPurchase(t, store)

	policy, err := domain.NewPolicy()
	if err != nil {
		t.Fatal(err)
	}

	service := operations.New(store, features.NewSnapshot(nil), policy)

	purchase, err = service.Submit(ctx, domain.Principal{
		ActorID: "alice", TenantID: "northstar", Role: domain.Member,
	}, purchase.ID)
	if err != nil {
		t.Fatal(err)
	}

	barrier := &lifecycleBarrier{
		Store:   store,
		loaded:  make(chan struct{}, 2),
		release: make(chan struct{}),
	}
	service = operations.New(barrier, features.NewSnapshot(nil), policy)
	results := make(chan error, 2)

	for _, decision := range []domain.Status{domain.Approved, domain.Rejected} {
		go func() {
			_, err := service.Decide(ctx, domain.Principal{
				ActorID: "bob", TenantID: "northstar", Role: domain.Approver,
			}, purchase.ID, decision)
			results <- err
		}()
	}

	for range 2 {
		select {
		case <-barrier.loaded:
		case <-ctx.Done():
			t.Fatal("operations did not load snapshots", ctx.Err())
		}
	}

	close(barrier.release)

	wins, conflicts := 0, 0

	for range 2 {
		select {
		case err := <-results:
			if err == nil {
				wins++
			} else if errors.Is(err, domain.ErrConflict) {
				conflicts++
			} else {
				t.Fatal(err)
			}
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}

	if wins != 1 || conflicts != 1 {
		t.Fatalf("wins=%d conflicts=%d", wins, conflicts)
	}

	saved, err := store.Purchase(ctx, "northstar", purchase.ID)
	if err != nil || saved.Version != 3 ||
		(saved.Status != domain.Approved && saved.Status != domain.Rejected) {
		t.Fatalf("invalid saved transition: %+v %v", saved, err)
	}

	events, err := store.Audit(ctx, "northstar", "")
	if err != nil || len(events) != 3 {
		t.Fatalf("stale decision appended audit: %d %v", len(events), err)
	}

	var count int
	if err := store.pool.QueryRow(ctx, "SELECT count(*) FROM pluto_jobs").
		Scan(&count); err != nil ||
		count != 3 {
		t.Fatalf("stale decision enqueued delivery: %d %v", count, err)
	}
}
