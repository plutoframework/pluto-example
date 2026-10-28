package operations

import (
	"context"
	"time"

	"github.com/plutoframework/pluto-example/internal/domain"

	"github.com/plutoframework/pluto-core/cache"
	rbac "github.com/plutoframework/pluto-rbac"
)

type OrganizationReader interface {
	Organization(context.Context, string) (domain.Organization, error)
}

type OrganizationService struct {
	reader      OrganizationReader
	loader      *cache.Loader[domain.Organization]
	policy      *rbac.Policy
	environment string
}

func NewOrganizationService(
	reader OrganizationReader,
	store cache.Store,
	policy *rbac.Policy,
	environment string,
) (*OrganizationService, error) {
	if reader == nil || policy == nil {
		return nil, domain.ErrInvalid
	}

	if _, err := organizationKey(environment, "validation"); err != nil {
		return nil, err
	}

	loader, err := cache.NewLoader[domain.Organization](
		store,
		cache.DefaultLoaderOptions(),
	)
	if err != nil {
		return nil, err
	}

	return &OrganizationService{
		reader:      reader,
		loader:      loader,
		policy:      policy,
		environment: environment,
	}, nil
}

func organizationKey(environment, tenant string) (string, error) {
	return cache.Key(cache.Scope{
		App:         "pluto-example",
		Environment: environment,
		Namespace:   "organization-profile",
		Version:     "1",
		TenantID:    tenant,
	}, "display")
}

// Organization caches display metadata only. Middleware resolves current membership
// and this operation checks permission on EVERY call, including cache hits.
func (s *OrganizationService) Organization(
	ctx context.Context,
	principal domain.Principal,
) (domain.Organization, error) {
	if !principal.Verified() {
		return domain.Organization{}, domain.ErrUnauthenticated
	}

	if !principal.Can(s.policy, domain.OrganizationRead) {
		return domain.Organization{}, domain.ErrForbidden
	}

	key, err := organizationKey(s.environment, principal.TenantID)
	if err != nil {
		return domain.Organization{}, err
	}

	value, err := s.loader.Remember(
		ctx,
		key,
		time.Minute,
		func(load context.Context) (domain.Organization, error) {
			return s.reader.Organization(load, principal.TenantID)
		},
	)
	if err != nil {
		return domain.Organization{}, err
	}

	if value.ID != principal.TenantID {
		return domain.Organization{}, domain.ErrNotFound
	}

	return value, nil
}
