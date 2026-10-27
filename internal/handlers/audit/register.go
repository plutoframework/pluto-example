package audit

import (
	"context"
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/paths"
	"github.com/plutoframework/pluto-example/internal/pluto"
)

type Operations interface {
	Audit(context.Context, domain.Principal) ([]domain.AuditEvent, error)
}

func Register(registrar *pluto.Registrar, operations Operations) {
	registrar.Handle(
		paths.ListAudit,
		func(writer http.ResponseWriter, request *http.Request) {
			events, err := operations.Audit(
				request.Context(),
				pluto.Principal(request),
			)
			if err != nil {
				pluto.Error(writer, request, err)

				return
			}

			result := make([]api.AuditEvent, 0, len(events))
			for _, event := range events {
				result = append(result, api.AuditEvent{
					ID:             event.ID,
					OrganizationID: event.TenantID,
					ActorID:        event.ActorID,
					Action:         event.Action,
					ResourceID:     event.ResourceID,
					Outcome:        event.Outcome,
					OccurredAt:     event.OccurredAt,
				})
			}

			pluto.JSON(
				writer,
				paths.ListAudit.SuccessStatus,
				api.AuditResponse{Data: result},
			)
		},
	)
}
