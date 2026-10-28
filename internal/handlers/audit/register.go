package audit

import (
	"context"
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/paths"
	"github.com/plutoframework/pluto-example/internal/pluto"

	"github.com/plutoframework/pluto-core/pagination"
)

type Operations interface {
	AuditPage(
		context.Context,
		domain.Principal,
		pagination.Request,
	) (pagination.Page[domain.AuditEvent], error)
}

func Register(registrar *pluto.Registrar, operations Operations) {
	registrar.Handle(
		paths.ListAudit,
		func(writer http.ResponseWriter, request *http.Request) {
			page, err := pluto.Pagination(request, paths.ListAudit.ID)
			if err != nil {
				pluto.Error(writer, request, err)

				return
			}

			events, err := operations.AuditPage(
				request.Context(),
				pluto.Principal(request),
				page,
			)
			if err != nil {
				pluto.Error(writer, request, err)

				return
			}

			result := make([]api.AuditEvent, 0, len(events.Data))
			for _, event := range events.Data {
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
				api.AuditResponse{Data: result, Page: events.Page},
			)
		},
	)
}
