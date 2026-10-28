package memory

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/plutoframework/pluto-example/internal/domain"

	"github.com/plutoframework/pluto-core/pagination"
)

func TestBoundedPagesRetainVisibilityAndStableTies(t *testing.T) {
	ctx := context.Background()
	store := New(
		[]domain.Membership{
			{TenantID: "one", ActorID: "a"},
			{TenantID: "one", ActorID: "b"},
			{TenantID: "two", ActorID: "c"},
		},
	)
	now := time.Date(2026, 10, 27, 1, 0, 0, 0, time.UTC)

	for i := range 12 {
		id := fmt.Sprintf("%02d", i)

		actor := "alice"
		if i%2 == 1 {
			actor = "bob"
		}

		tenant := "one"
		if i >= 10 {
			tenant = "two"
		}

		p := domain.Purchase{ID: id, TenantID: tenant, CreatorID: actor}

		event := domain.AuditEvent{
			ID:         id,
			TenantID:   tenant,
			ActorID:    actor,
			OccurredAt: now,
		}
		if err := store.SavePurchase(ctx, p, 0, event); err != nil {
			t.Fatal(err)
		}
	}

	binding := pagination.Binding(nil, "purchases.list", "alice", "one")
	request, _ := pagination.Parse(url.Values{"limit": {"2"}}, binding)
	got := []string{}

	for {
		rows, err := store.PurchasePage(ctx, "one", "alice", request)
		if err != nil || len(rows) > 3 {
			t.Fatal(rows, err)
		}

		page, err := pagination.Finish(
			request,
			rows,
			func(p domain.Purchase) pagination.Key { return pagination.Key{ID: p.ID} },
		)
		if err != nil {
			t.Fatal(err)
		}

		for _, p := range page.Data {
			got = append(got, p.ID)
		}

		if !page.Page.HasMore {
			break
		}

		request, err = pagination.Parse(
			url.Values{"limit": {"2"}, "cursor": {page.Page.NextCursor}},
			binding,
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	if fmt.Sprint(got) != "[00 02 04 06 08]" {
		t.Fatal("tenant/owner/boundary lost", got)
	}

	request, _ = pagination.Parse(
		url.Values{"limit": {"2"}},
		pagination.Binding(nil, "audit", "alice", "one"),
	)

	first, err := store.AuditPage(ctx, "one", "alice", request)
	if err != nil || len(first) != 3 {
		t.Fatal(first, err)
	}

	page, err := pagination.Finish(request, first, domain.AuditKey)
	if err != nil {
		t.Fatal(err)
	}

	next, err := pagination.Parse(
		url.Values{"limit": {"2"}, "cursor": {page.Page.NextCursor}},
		pagination.Binding(nil, "audit", "alice", "one"),
	)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := store.AuditPage(ctx, "one", "alice", next)
	if err != nil || len(rows) != 3 || rows[0].ID != "04" ||
		rows[2].ID != "08" {
		t.Fatal("same timestamp tie-breaker lost", rows, err)
	}

	request, _ = pagination.Parse(
		url.Values{"limit": {"1"}},
		pagination.Binding(nil, "members", "one"),
	)

	members, err := store.MembershipPage(ctx, "one", request)
	if err != nil || len(members) != 2 || members[0].ActorID != "a" ||
		members[1].ActorID != "b" {
		t.Fatal(members, err)
	}

	if _, err := store.PurchasePage(
		ctx,
		"one",
		"alice",
		pagination.Request{Limit: 101},
	); err == nil {
		t.Fatal("unbounded repository request")
	}
}
