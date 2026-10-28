package postgres

import (
	"context"

	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/storage/postgres/db"

	"github.com/jackc/pgx/v5/pgtype"
)

func appendAudit(
	ctx context.Context,
	queries *db.Queries,
	event domain.AuditEvent,
) error {
	return queries.AppendAudit(ctx, db.AppendAuditParams{
		TenantID:   event.TenantID,
		ID:         event.ID,
		ActorID:    event.ActorID,
		Action:     event.Action,
		ResourceID: event.ResourceID,
		Outcome:    event.Outcome,
		OccurredAt: pgtype.Timestamptz{
			Time:  event.OccurredAt,
			Valid: true,
		},
	})
}

func (s *Store) SavePurchase(
	ctx context.Context,
	purchase domain.Purchase,
	expected int64,
	event domain.AuditEvent,
) error {
	if event.TenantID != purchase.TenantID {
		return domain.ErrInvalid
	}

	transaction, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer transaction.Rollback(ctx)

	queries := s.queries.WithTx(transaction)
	if expected == 0 {
		err = queries.CreatePurchase(ctx, db.CreatePurchaseParams{
			TenantID:    purchase.TenantID,
			ID:          purchase.ID,
			CreatorID:   purchase.CreatorID,
			Title:       purchase.Title,
			AmountCents: purchase.AmountCents,
			Status:      string(purchase.Status),
			Version:     purchase.Version,
			CreatedAt: pgtype.Timestamptz{
				Time:  purchase.CreatedAt,
				Valid: true,
			},
			UpdatedAt: timestamp(purchase),
		})
	} else {
		var affected int64

		affected, err = queries.UpdatePurchase(ctx, db.UpdatePurchaseParams{
			TenantID:  purchase.TenantID,
			ID:        purchase.ID,
			Status:    string(purchase.Status),
			Version:   purchase.Version,
			UpdatedAt: timestamp(purchase),
			Version_2: expected,
		})
		if err == nil && affected != 1 {
			return domain.ErrConflict
		}
	}

	if err != nil {
		return translate(err)
	}

	if err := appendAudit(ctx, queries, event); err != nil {
		return translate(err)
	}

	if err := s.enqueueAudit(ctx, transaction, event); err != nil {
		return translate(err)
	}

	return translate(transaction.Commit(ctx))
}

func (s *Store) ChangeRole(
	ctx context.Context,
	tenant, actor string,
	role domain.Role,
	event domain.AuditEvent,
) error {
	if event.TenantID != tenant {
		return domain.ErrInvalid
	}

	transaction, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer transaction.Rollback(ctx)

	queries := s.queries.WithTx(transaction)

	affected, err := queries.ChangeRole(ctx, db.ChangeRoleParams{
		TenantID: tenant,
		ActorID:  actor,
		Role:     string(role),
	})
	if err != nil {
		return translate(err)
	}

	if affected != 1 {
		return domain.ErrNotFound
	}

	if err := appendAudit(ctx, queries, event); err != nil {
		return translate(err)
	}

	if err := s.enqueueAudit(ctx, transaction, event); err != nil {
		return translate(err)
	}

	return translate(transaction.Commit(ctx))
}
