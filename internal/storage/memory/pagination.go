package memory

import (
	"context"
	"sort"

	"github.com/plutoframework/pluto-example/internal/domain"

	"github.com/plutoframework/pluto-core/pagination"
)

// Keep only the next limit+1 rows while scanning; memory never scales with the
// collection size. A real database uses indexed predicates and LIMIT instead.
func insertBounded[T any](
	rows []T,
	value T,
	max int,
	less func(T, T) bool,
) []T {
	index := sort.Search(
		len(rows),
		func(i int) bool { return !less(rows[i], value) },
	)
	if index >= max {
		return rows
	}

	if len(rows) < max {
		rows = append(rows, value)
	}

	copy(rows[index+1:], rows[index:len(rows)-1])

	rows[index] = value

	return rows
}

func (s *Store) PurchasePage(
	ctx context.Context,
	tenant, actor string,
	request pagination.Request,
) ([]domain.Purchase, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	if request.After.Sort != "" {
		return nil, pagination.ErrInvalid
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := []domain.Purchase{}

	for _, v := range s.purchases {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if v.TenantID != tenant || (actor != "" && v.CreatorID != actor) ||
			v.ID <= request.After.ID {
			continue
		}

		rows = insertBounded(
			rows,
			v,
			request.FetchLimit(),
			func(a, b domain.Purchase) bool { return a.ID < b.ID },
		)
	}

	return rows, nil
}

func (s *Store) MembershipPage(
	ctx context.Context,
	tenant string,
	request pagination.Request,
) ([]domain.Membership, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	if request.After.Sort != "" {
		return nil, pagination.ErrInvalid
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := []domain.Membership{}

	for _, v := range s.memberships {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if v.TenantID != tenant || v.ActorID <= request.After.ID {
			continue
		}

		rows = insertBounded(
			rows,
			v,
			request.FetchLimit(),
			func(a, b domain.Membership) bool { return a.ActorID < b.ActorID },
		)
	}

	return rows, nil
}

func (s *Store) AuditPage(
	ctx context.Context,
	tenant, actor string,
	request pagination.Request,
) ([]domain.AuditEvent, error) {
	boundary, err := domain.AuditBoundary(request)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := []domain.AuditEvent{}
	less := func(a, b domain.AuditEvent) bool {
		return a.OccurredAt.Before(b.OccurredAt) ||
			(a.OccurredAt.Equal(b.OccurredAt) && a.ID < b.ID)
	}

	for _, v := range s.events {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if v.TenantID != tenant || (actor != "" && v.ActorID != actor) {
			continue
		}

		if request.After.ID != "" &&
			(v.OccurredAt.Before(boundary) || (v.OccurredAt.Equal(boundary) && v.ID <= request.After.ID)) {
			continue
		}

		rows = insertBounded(rows, v, request.FetchLimit(), less)
	}

	return rows, nil
}
