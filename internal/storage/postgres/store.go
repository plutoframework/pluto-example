// Package postgres implements the repository using concrete pgx and sqlc queries.
package postgres

import (
	"context"
	"errors"

	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/storage/postgres/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func Open(ctx context.Context, url string) (*Store, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()

		return nil, err
	}

	return &Store{
		pool:    pool,
		queries: db.New(pool),
	}, nil
}

// Pool exposes the explicit composition boundary for shared durable adapters.
func (s *Store) Pool() *pgxpool.Pool { return s.pool }

func (s *Store) Close() { s.pool.Close() }

func translate(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return domain.ErrConflict
	}

	return err
}

func (s *Store) Membership(
	ctx context.Context,
	tenant, actor string,
) (domain.Membership, error) {
	value, err := s.queries.GetMembership(ctx, db.GetMembershipParams{
		TenantID: tenant,
		ActorID:  actor,
	})

	return membershipModel(value), translate(err)
}

func (s *Store) Memberships(
	ctx context.Context,
	tenant string,
) ([]domain.Membership, error) {
	values, err := s.queries.ListMemberships(ctx, tenant)
	if err != nil {
		return nil, translate(err)
	}

	result := make([]domain.Membership, 0, len(values))
	for _, value := range values {
		result = append(result, membershipModel(value))
	}

	return result, nil
}

func membershipModel(value db.Membership) domain.Membership {
	return domain.Membership{
		TenantID:    value.TenantID,
		ActorID:     value.ActorID,
		DisplayName: value.DisplayName,
		Role:        domain.Role(value.Role),
	}
}

func (s *Store) Purchase(
	ctx context.Context,
	tenant, id string,
) (domain.Purchase, error) {
	value, err := s.queries.GetPurchase(ctx, db.GetPurchaseParams{
		TenantID: tenant,
		ID:       id,
	})

	return purchaseModel(value), translate(err)
}

func (s *Store) Purchases(
	ctx context.Context,
	tenant string,
) ([]domain.Purchase, error) {
	values, err := s.queries.ListPurchases(ctx, tenant)
	if err != nil {
		return nil, translate(err)
	}

	result := make([]domain.Purchase, 0, len(values))
	for _, value := range values {
		result = append(result, purchaseModel(value))
	}

	return result, nil
}

func purchaseModel(value db.Purchase) domain.Purchase {
	return domain.Purchase{
		ID:          value.ID,
		TenantID:    value.TenantID,
		CreatorID:   value.CreatorID,
		Title:       value.Title,
		AmountCents: value.AmountCents,
		Status:      domain.Status(value.Status),
		Version:     value.Version,
		CreatedAt:   value.CreatedAt.Time,
		UpdatedAt:   value.UpdatedAt.Time,
	}
}

func (s *Store) Audit(
	ctx context.Context,
	tenant, actor string,
) ([]domain.AuditEvent, error) {
	values, err := s.queries.ListAudit(ctx, db.ListAuditParams{
		TenantID:    tenant,
		ActorFilter: actor,
	})
	if err != nil {
		return nil, translate(err)
	}

	result := make([]domain.AuditEvent, 0, len(values))
	for _, value := range values {
		result = append(result, domain.AuditEvent{
			ID:         value.ID,
			TenantID:   value.TenantID,
			ActorID:    value.ActorID,
			Action:     value.Action,
			ResourceID: value.ResourceID,
			Outcome:    value.Outcome,
			OccurredAt: value.OccurredAt.Time,
		})
	}

	return result, nil
}

func timestamp(value domain.Purchase) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  value.UpdatedAt,
		Valid: true,
	}
}
