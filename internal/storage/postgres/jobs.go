package postgres

import (
	"context"

	"github.com/plutoframework/pluto-example/internal/background"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/storage/postgres/db"

	"github.com/jackc/pgx/v5"
	jobs "github.com/plutoframework/pluto-jobs"
)

func (s *Store) enqueueAudit(
	ctx context.Context,
	tx pgx.Tx,
	event domain.AuditEvent,
) error {
	request, err := jobs.NewRequest(
		background.AuditDelivery,
		background.AuditReference{
			EventID: event.ID,
		},
		jobs.Options{TenantID: &event.TenantID},
	)
	if err != nil {
		return err
	}

	return s.outbox.EnqueueTx(ctx, tx, request)
}

// ArchiveAudit is a local demonstration sink, not an external SIEM integration.
// The tenant/event key makes repeated delivery safe after an acknowledgement loss.
func (s *Store) ArchiveAudit(
	ctx context.Context,
	jobID, tenant, eventID string,
) error {
	_, err := s.queries.ArchiveAudit(ctx, db.ArchiveAuditParams{
		JobID:    jobID,
		TenantID: tenant,
		EventID:  eventID,
	})

	return translate(err)
}
