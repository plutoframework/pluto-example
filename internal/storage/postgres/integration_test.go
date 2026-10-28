package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/features"
	"github.com/plutoframework/pluto-example/internal/operations"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func isolatedStore(t *testing.T) *Store {
	t.Helper()

	connectionURL := os.Getenv("PLUTO_EXAMPLE_TEST_DATABASE_URL")
	if connectionURL == "" {
		t.Skip(
			"SQL integration not run: set PLUTO_EXAMPLE_TEST_DATABASE_URL to a dedicated test database",
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	admin, err := pgxpool.New(ctx, connectionURL)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(admin.Close)

	var suffix [6]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		t.Fatal(err)
	}

	schemaName := "pluto_test_" + hex.EncodeToString(suffix[:])

	identifier := pgx.Identifier{schemaName}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+identifier); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if _, err := admin.Exec(
			cleanup,
			"DROP SCHEMA "+identifier+" CASCADE",
		); err != nil {
			t.Errorf("cleanup isolated schema: %v", err)
		}
	})

	parsed, err := url.Parse(connectionURL)
	if err != nil {
		t.Fatal(err)
	}

	parameters := parsed.Query()
	parameters.Set("search_path", schemaName)

	parsed.RawQuery = parameters.Encode()

	store, err := Open(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(store.Close)

	for _, name := range []string{
		"migrations/001_initial.sql",
		"migrations/002_auth_sessions.sql",
		"migrations/003_idempotency.sql",
		"fixtures.sql",
	} {
		source, err := os.ReadFile(filepath.Join("..", "..", "..", "db", name))
		if err != nil {
			t.Fatal(err)
		}

		if _, err := store.pool.Exec(ctx, string(source)); err != nil {
			t.Fatalf("apply %s: %v", name, err)
		}
	}

	return store
}

func TestPostgresTenantIsolationAtomicAuditAndConcurrency(t *testing.T) {
	store := isolatedStore(t)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	policy, err := domain.NewPolicy()
	if err != nil {
		t.Fatal(err)
	}

	service := operations.New(store, features.NewSnapshot(nil), policy)
	alice := domain.Principal{
		ActorID:  "alice",
		TenantID: "northstar",
		Role:     domain.Member,
	}

	purchase, err := service.Create(ctx, alice, "Calibration service", 48000)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.Purchase(
		ctx,
		"harbor",
		purchase.ID,
	); !errors.Is(
		err,
		domain.ErrNotFound,
	) {
		t.Fatalf("cross-tenant SQL read: %v", err)
	}

	values, err := store.Purchases(ctx, "harbor")
	if err != nil || len(values) != 0 {
		t.Fatalf("cross-tenant SQL list: %+v %v", values, err)
	}

	events, err := store.Audit(ctx, "northstar", "")
	if err != nil || len(events) != 1 {
		t.Fatalf("create/audit not atomic: %+v %v", events, err)
	}

	otherEvents, err := store.Audit(ctx, "harbor", "")
	if err != nil || len(otherEvents) != 0 {
		t.Fatalf("cross-tenant SQL audit: %+v %v", otherEvents, err)
	}

	// Force the audit insert to fail after the purchase insert. The duplicate audit
	// primary key must roll the purchase insert back, not leave an unaudited row.
	copyPurchase := purchase

	copyPurchase.ID = "must-roll-back"
	if err := store.SavePurchase(
		ctx,
		copyPurchase,
		0,
		events[0],
	); !errors.Is(
		err,
		domain.ErrConflict,
	) {
		t.Fatalf("expected duplicate audit conflict: %v", err)
	}

	if _, err := store.Purchase(
		ctx,
		"northstar",
		copyPurchase.ID,
	); !errors.Is(
		err,
		domain.ErrNotFound,
	) {
		t.Fatalf("failed audit left a purchase: %v", err)
	}

	// The same failure during a transition must also roll back its UPDATE.
	updated := purchase
	updated.Status = domain.Submitted

	updated.Version = 2
	if err := store.SavePurchase(
		ctx,
		updated,
		1,
		events[0],
	); !errors.Is(
		err,
		domain.ErrConflict,
	) {
		t.Fatalf("expected transition rollback: %v", err)
	}

	unchanged, err := store.Purchase(ctx, "northstar", purchase.ID)
	if err != nil || unchanged.Status != domain.Draft ||
		unchanged.Version != 1 {
		t.Fatalf("failed audit changed purchase: %+v %v", unchanged, err)
	}

	results := make(chan error, 2)

	for index, status := range []domain.Status{
		domain.Submitted,
		domain.Rejected,
	} {
		candidate := updated
		candidate.Status = status
		event := events[0]
		event.ID = "concurrent-" + string(rune('a'+index))

		event.Action = "test.concurrent_write"
		go func() { results <- store.SavePurchase(ctx, candidate, 1, event) }()
	}

	successes, conflicts := 0, 0

	for index := 0; index < 2; index++ {
		result := <-results
		if result == nil {
			successes++
		} else if errors.Is(result, domain.ErrConflict) {
			conflicts++
		} else {
			t.Fatal(result)
		}
	}

	if successes != 1 || conflicts != 1 {
		t.Fatalf("winners=%d conflicts=%d", successes, conflicts)
	}

	events, err = store.Audit(ctx, "northstar", "")
	if err != nil || len(events) != 2 {
		t.Fatalf("stale or failed write appended audit: %+v %v", events, err)
	}

	// Membership predicates and role/audit transaction use the same tenant boundary.
	event := events[0]
	event.ID = "role-event"

	event.Action = "membership.role_changed"
	if err := store.ChangeRole(
		ctx,
		"northstar",
		"dana",
		domain.Member,
		event,
	); !errors.Is(
		err,
		domain.ErrNotFound,
	) {
		t.Fatalf("cross-tenant membership changed: %v", err)
	}

	if err := store.ChangeRole(
		ctx,
		"northstar",
		"bob",
		domain.Member,
		events[0],
	); !errors.Is(
		err,
		domain.ErrConflict,
	) {
		t.Fatalf("expected role audit rollback: %v", err)
	}

	membership, err := store.Membership(ctx, "northstar", "bob")
	if err != nil || membership.Role != domain.Approver {
		t.Fatalf("failed audit changed role: %+v %v", membership, err)
	}
}
