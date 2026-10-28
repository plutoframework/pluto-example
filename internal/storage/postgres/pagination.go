package postgres

import (
	"context"

	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/storage/postgres/db"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/plutoframework/pluto-core/pagination"
)

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

	values, err := s.queries.PagePurchases(
		ctx,
		db.PagePurchasesParams{
			TenantID:    tenant,
			ActorFilter: actor,
			AfterID:     request.After.ID,
			FetchLimit:  int32(request.FetchLimit()),
		},
	)
	if err != nil {
		return nil, translate(err)
	}

	rows := make([]domain.Purchase, 0, len(values))
	for _, v := range values {
		rows = append(rows, purchaseModel(v))
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

	values, err := s.queries.PageMemberships(
		ctx,
		db.PageMembershipsParams{
			TenantID:   tenant,
			AfterID:    request.After.ID,
			FetchLimit: int32(request.FetchLimit()),
		},
	)
	if err != nil {
		return nil, translate(err)
	}

	rows := make([]domain.Membership, 0, len(values))
	for _, v := range values {
		rows = append(rows, membershipModel(v))
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

	values, err := s.queries.PageAudit(
		ctx,
		db.PageAuditParams{
			TenantID:    tenant,
			ActorFilter: actor,
			AfterID:     request.After.ID,
			AfterTime:   pgtype.Timestamptz{Time: boundary, Valid: true},
			FetchLimit:  int32(request.FetchLimit()),
		},
	)
	if err != nil {
		return nil, translate(err)
	}

	rows := make([]domain.AuditEvent, 0, len(values))
	for _, v := range values {
		rows = append(
			rows,
			domain.AuditEvent{
				ID:         v.ID,
				TenantID:   v.TenantID,
				ActorID:    v.ActorID,
				Action:     v.Action,
				ResourceID: v.ResourceID,
				Outcome:    v.Outcome,
				OccurredAt: v.OccurredAt.Time,
			},
		)
	}

	return rows, nil
}
