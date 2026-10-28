package postgres

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/plutoframework/pluto-example/internal/domain"

	"github.com/plutoframework/pluto-core/pagination"
)

func TestPostgresKeysetVisibilityAndTies(t *testing.T) {
	store := isolatedStore(t)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	now := time.Date(2026, 10, 27, 1, 0, 0, 0, time.UTC)

	for i := range 6 {
		actor := "alice"
		if i%2 == 1 {
			actor = "bob"
		}

		id := fmt.Sprintf("page-%02d", i)
		p := domain.Purchase{
			ID:          id,
			TenantID:    "northstar",
			CreatorID:   actor,
			Title:       "Lab supplies",
			AmountCents: 100,
			Status:      domain.Draft,
			Version:     1,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		event := domain.AuditEvent{
			ID:         id,
			TenantID:   "northstar",
			ActorID:    actor,
			Action:     "purchase.created",
			ResourceID: id,
			Outcome:    "success",
			OccurredAt: now,
		}
		if err := store.SavePurchase(ctx, p, 0, event); err != nil {
			t.Fatal(err)
		}
	}

	binding := pagination.Binding(nil, "list", "northstar", "alice")
	request, _ := pagination.Parse(url.Values{"limit": {"1"}}, binding)

	rows, err := store.PurchasePage(ctx, "northstar", "alice", request)
	if err != nil || len(rows) != 2 || rows[0].ID != "page-00" ||
		rows[1].ID != "page-02" {
		t.Fatal(rows, err)
	}

	first, err := pagination.Finish(
		request,
		rows,
		func(p domain.Purchase) pagination.Key { return pagination.Key{ID: p.ID} },
	)
	if err != nil {
		t.Fatal(err)
	}

	request, err = pagination.Parse(
		url.Values{"limit": {"2"}, "cursor": {first.Page.NextCursor}},
		binding,
	)
	if err != nil {
		t.Fatal(err)
	}

	rows, err = store.PurchasePage(ctx, "northstar", "alice", request)
	if err != nil || len(rows) != 2 || rows[0].ID != "page-02" ||
		rows[1].ID != "page-04" {
		t.Fatal(rows, err)
	}

	hidden, err := store.PurchasePage(ctx, "harbor", "alice", request)
	if err != nil || len(hidden) != 0 {
		t.Fatal("tenant predicate lost", hidden, err)
	}

	request, _ = pagination.Parse(
		url.Values{"limit": {"1"}},
		pagination.Binding(nil, "audit", "northstar", "alice"),
	)

	events, err := store.AuditPage(ctx, "northstar", "alice", request)
	if err != nil || len(events) != 2 {
		t.Fatal(events, err)
	}

	audit, err := pagination.Finish(request, events, domain.AuditKey)
	if err != nil {
		t.Fatal(err)
	}

	request, err = pagination.Parse(
		url.Values{"limit": {"1"}, "cursor": {audit.Page.NextCursor}},
		pagination.Binding(nil, "audit", "northstar", "alice"),
	)
	if err != nil {
		t.Fatal(err)
	}

	events, err = store.AuditPage(ctx, "northstar", "alice", request)
	if err != nil || len(events) != 2 || events[0].ID != "page-02" {
		t.Fatal("audit tie-breaker lost", events, err)
	}

	if _, err := store.PurchasePage(
		ctx,
		"northstar",
		"alice",
		pagination.Request{Limit: 101},
	); err == nil {
		t.Fatal("unbounded SQL query")
	}
}
