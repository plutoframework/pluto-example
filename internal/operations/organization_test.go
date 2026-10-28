package operations

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/plutoframework/pluto-example/internal/domain"

	"github.com/plutoframework/pluto-core/cache"
	"github.com/plutoframework/pluto-core/cache/memory"
)

type organizationReader struct{ calls int }

func (r *organizationReader) Organization(
	_ context.Context,
	tenant string,
) (domain.Organization, error) {
	r.calls++
	return domain.Organization{ID: tenant, Name: tenant + " display"}, nil
}

type failedCache struct{}

func (failedCache) Get(context.Context, string) ([]byte, bool, error) {
	return nil, false, cache.ErrUnavailable
}

func (failedCache) Set(context.Context, string, []byte, time.Duration) error {
	return cache.ErrUnavailable
}

func (failedCache) Delete(
	context.Context,
	string,
) error {
	return cache.ErrUnavailable
}

func TestOrganizationCacheScopeAndAuthorization(t *testing.T) {
	policy, _ := domain.NewPolicy()
	store, _ := memory.New(memory.DefaultConfig())
	reader := &organizationReader{}

	service, err := NewOrganizationService(reader, store, policy, "test")
	if err != nil {
		t.Fatal(err)
	}

	principal := domain.Principal{
		ActorID:  "alice",
		TenantID: "northstar",
		Role:     domain.Member,
	}
	for range 2 {
		value, err := service.Organization(context.Background(), principal)
		if err != nil || value.ID != "northstar" {
			t.Fatal(value, err)
		}
	}

	if reader.calls != 1 {
		t.Fatal("cache hit reloaded", reader.calls)
	}

	principal.TenantID = "harbor"

	value, err := service.Organization(context.Background(), principal)
	if err != nil || value.ID != "harbor" || reader.calls != 2 {
		t.Fatal("tenant data crossed", value, err)
	}

	principal.Role = "revoked"
	if _, err := service.Organization(
		context.Background(),
		principal,
	); !errors.Is(err, domain.ErrForbidden) ||
		reader.calls != 2 {
		t.Fatal("cache bypassed permission", err)
	}

	principal.Role, principal.ActorID = domain.Member, ""
	if _, err := service.Organization(
		context.Background(),
		principal,
	); !errors.Is(
		err,
		domain.ErrUnauthenticated,
	) {
		t.Fatal("cache bypassed identity", err)
	}

	other, _ := NewOrganizationService(
		reader,
		store,
		policy,
		"other-environment",
	)

	principal.ActorID = "alice"
	if _, err := other.Organization(
		context.Background(),
		principal,
	); err != nil ||
		reader.calls != 3 {
		t.Fatal("environment collision", err)
	}
}

func TestOrganizationCacheOutageAndInvalidation(t *testing.T) {
	policy, _ := domain.NewPolicy()
	reader := &organizationReader{}
	service, _ := NewOrganizationService(reader, failedCache{}, policy, "test")

	principal := domain.Principal{
		ActorID:  "alice",
		TenantID: "northstar",
		Role:     domain.Member,
	}
	if value, err := service.Organization(
		context.Background(),
		principal,
	); err != nil || value.ID != "northstar" ||
		reader.calls != 1 {
		t.Fatal("cache outage broke database path", value, err)
	}

	store, _ := memory.New(memory.DefaultConfig())
	service, _ = NewOrganizationService(reader, store, policy, "test")
	_, _ = service.Organization(context.Background(), principal)

	key, _ := organizationKey("test", "northstar")
	if err := store.Delete(context.Background(), key); err != nil {
		t.Fatal(err)
	}

	_, _ = service.Organization(context.Background(), principal)

	if reader.calls != 3 {
		t.Fatal("delete did not reload", reader.calls)
	}

	if err := cache.SetJSON(
		context.Background(),
		store,
		key,
		domain.Organization{ID: "harbor", Name: "wrong tenant"},
		time.Minute,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := service.Organization(
		context.Background(),
		principal,
	); !errors.Is(
		err,
		domain.ErrNotFound,
	) {
		t.Fatal("poisoned scope returned", err)
	}
}
