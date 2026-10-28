// Package background defines explicit application job contracts and handlers.
package background

import (
	"context"
	"errors"
	"strings"

	"github.com/plutoframework/pluto-example/internal/domain"

	jobs "github.com/plutoframework/pluto-jobs"
)

const AuditDelivery jobs.Kind = "audit.deliver"

// AuditReference contains no credentials or audit body. The worker loads the
// canonical event with the exact tenant, never a request-supplied identity.
type AuditReference struct {
	EventID string `json:"event_id"`
}

type AuditSink interface {
	ArchiveAudit(context.Context, string, string, string) error
}

func AuditHandler(sink AuditSink) jobs.Handler {
	return func(ctx context.Context, job jobs.Job) error {
		payload, err := jobs.Decode[AuditReference](job)
		if err != nil || job.Kind != AuditDelivery || job.TenantID == nil ||
			payload.EventID == "" || len(payload.EventID) > 128 ||
			strings.TrimSpace(payload.EventID) != payload.EventID {
			return jobs.Permanent(jobs.ErrInvalid)
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if sink == nil {
			return errors.New("audit delivery requires a sink")
		}

		err = sink.ArchiveAudit(ctx, job.ID, *job.TenantID, payload.EventID)
		if errors.Is(err, domain.ErrNotFound) {
			return jobs.Permanent(domain.ErrNotFound)
		}

		return err
	}
}

func NewWorker(
	store jobs.Store,
	sink AuditSink,
	config jobs.Config,
) (*jobs.Worker, error) {
	if sink == nil {
		return nil, jobs.ErrInvalid
	}

	return jobs.NewWorker(store, map[jobs.Kind]jobs.Handler{
		AuditDelivery: AuditHandler(sink),
	}, config)
}
