package postgres

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/plutoframework/pluto-example/internal/background"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/features"
	"github.com/plutoframework/pluto-example/internal/operations"

	jobs "github.com/plutoframework/pluto-jobs"
)

func createJobPurchase(t *testing.T, store *Store) domain.Purchase {
	t.Helper()

	policy, err := domain.NewPolicy()
	if err != nil {
		t.Fatal(err)
	}

	service := operations.New(store, features.NewSnapshot(nil), policy)

	purchase, err := service.Create(context.Background(), domain.Principal{
		ActorID: "alice", TenantID: "northstar", Role: domain.Member,
	}, "Background delivery", 5000)
	if err != nil {
		t.Fatal(err)
	}

	return purchase
}

func TestPostgresJobsAtomicBusinessEnqueue(t *testing.T) {
	store := isolatedStore(t)
	ctx := context.Background()
	purchase := createJobPurchase(t, store)

	var jobsBefore int
	if err := store.pool.QueryRow(ctx, "SELECT count(*) FROM pluto_jobs").
		Scan(&jobsBefore); err != nil ||
		jobsBefore != 1 {
		t.Fatal(
			"business write did not enqueue exactly one job",
			jobsBefore,
			err,
		)
	}

	// Force enqueue to fail AFTER the business mutation and audit append.
	if _, err := store.pool.Exec(
		ctx,
		"ALTER TABLE pluto_jobs ADD CONSTRAINT reject_delivery CHECK (kind <> 'audit.deliver') NOT VALID",
	); err != nil {
		t.Fatal(err)
	}

	events, err := store.Audit(ctx, "northstar", "")
	if err != nil || len(events) != 1 {
		t.Fatal(events, err)
	}

	event := events[0]
	event.ID = "enqueue-rollback"
	copy := purchase

	copy.ID = "must-not-commit"
	if err := store.SavePurchase(ctx, copy, 0, event); err == nil {
		t.Fatal("business write committed despite outbox failure")
	}

	if _, err := store.Purchase(
		ctx,
		"northstar",
		copy.ID,
	); !errors.Is(
		err,
		domain.ErrNotFound,
	) {
		t.Fatal("outbox failure left a purchase", err)
	}

	if err := store.ChangeRole(
		ctx,
		"northstar",
		"bob",
		domain.Member,
		event,
	); err == nil {
		t.Fatal("membership committed despite outbox failure")
	}

	member, err := store.Membership(ctx, "northstar", "bob")
	if err != nil || member.Role != domain.Approver {
		t.Fatal("membership rollback", member, err)
	}

	events, err = store.Audit(ctx, "northstar", "")
	if err != nil || len(events) != 1 {
		t.Fatal("outbox failure left audit rows", events, err)
	}

	var after int
	if err := store.pool.QueryRow(ctx, "SELECT count(*) FROM pluto_jobs").
		Scan(&after); err != nil ||
		after != jobsBefore {
		t.Fatal("outbox failure left work", after, err)
	}
}

type retryAuditSink struct {
	store *Store
	calls atomic.Int32
}

func (s *retryAuditSink) ArchiveAudit(
	ctx context.Context,
	job, tenant, event string,
) error {
	if s.calls.Add(1) == 1 {
		return errors.New("synthetic transient delivery failure")
	}

	return s.store.ArchiveAudit(ctx, job, tenant, event)
}

func TestPostgresJobsWorkerRetryAndTenant(t *testing.T) {
	store := isolatedStore(t)
	createJobPurchase(t, store)

	sink := &retryAuditSink{store: store}
	config := jobs.DefaultConfig()
	config.Concurrency = 1
	config.PollInterval = 5 * time.Millisecond
	config.RetryBase = 5 * time.Millisecond
	config.RetryMax = 20 * time.Millisecond

	worker, err := background.NewWorker(store.outbox, sink, config)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()

	t.Cleanup(func() {
		cancel()

		if err := <-done; err != nil && !errors.Is(err, context.Canceled) {
			t.Error(err)
		}
	})

	var (
		state, jobID string
		attempts     int
	)
	for {
		err := store.pool.QueryRow(ctx, "SELECT id, state, attempts FROM pluto_jobs").
			Scan(&jobID, &state, &attempts)
		if err != nil {
			t.Fatal(err)
		}

		if state == "succeeded" {
			break
		}

		select {
		case <-ctx.Done():
			t.Fatal("worker did not settle", state, attempts)
		case <-time.After(5 * time.Millisecond):
		}
	}

	if attempts != 2 || sink.calls.Load() != 2 {
		t.Fatal("retry accounting", attempts, sink.calls.Load())
	}

	events, err := store.Audit(ctx, "northstar", "")
	if err != nil || len(events) != 1 {
		t.Fatal(events, err)
	}

	// Simulate a provider success followed by lost acknowledgement and redelivery.
	for range 2 {
		if err := store.ArchiveAudit(
			ctx,
			jobID,
			"northstar",
			events[0].ID,
		); err != nil {
			t.Fatal(err)
		}
	}

	var count int
	if err := store.pool.QueryRow(ctx, "SELECT count(*) FROM audit_archive WHERE tenant_id='northstar'").
		Scan(&count); err != nil ||
		count != 1 {
		t.Fatal("duplicate delivery duplicated the effect", count, err)
	}

	if err := store.ArchiveAudit(
		ctx,
		jobs.NewID(),
		"harbor",
		events[0].ID,
	); !errors.Is(
		err,
		domain.ErrNotFound,
	) {
		t.Fatal("cross-tenant event delivered", err)
	}
}

func TestPostgresJobsRollbackAndCommitVisibility(t *testing.T) {
	store := isolatedStore(t)
	ctx := context.Background()

	event := domain.AuditEvent{
		ID:       "synthetic",
		TenantID: "northstar",
	}

	for _, commit := range []bool{false, true} {
		tx, err := store.pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if err := store.enqueueAudit(ctx, tx, event); err != nil {
			t.Fatal(err)
		}

		var requestID string
		if err := tx.QueryRow(ctx, "SELECT id FROM pluto_jobs").
			Scan(&requestID); err != nil {
			t.Fatal(err)
		}

		if _, err := store.outbox.Claim(
			ctx,
			time.Minute,
		); !errors.Is(
			err,
			jobs.ErrNoJob,
		) {
			t.Fatal("uncommitted work visible", err)
		}

		if commit {
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}

			claim, err := store.outbox.Claim(ctx, time.Minute)
			if err != nil || claim.Job.ID != requestID {
				t.Fatal("committed job missing", err)
			}
		} else {
			if err := tx.Rollback(ctx); err != nil {
				t.Fatal(err)
			}

			if _, err := store.outbox.Claim(
				ctx,
				time.Minute,
			); !errors.Is(
				err,
				jobs.ErrNoJob,
			) {
				t.Fatal("rollback left a job", err)
			}
		}
	}
}
