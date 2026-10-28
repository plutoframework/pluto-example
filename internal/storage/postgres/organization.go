package postgres

import (
	"context"

	"github.com/plutoframework/pluto-example/internal/domain"
)

func (s *Store) Organization(
	ctx context.Context,
	tenant string,
) (domain.Organization, error) {
	value, err := s.queries.Organization(ctx, tenant)
	return domain.Organization{ID: value.ID, Name: value.Name}, translate(err)
}
