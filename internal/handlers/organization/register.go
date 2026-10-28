package organization

import (
	"context"
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/paths"
	"github.com/plutoframework/pluto-example/internal/pluto"
)

type Operations interface {
	Organization(context.Context, domain.Principal) (domain.Organization, error)
}

func Register(registrar *pluto.Registrar, operations Operations) {
	registrar.Handle(
		paths.GetOrganization,
		func(writer http.ResponseWriter, request *http.Request) {
			value, err := operations.Organization(
				request.Context(),
				pluto.Principal(request),
			)
			if err != nil {
				pluto.Error(writer, request, err)

				return
			}

			pluto.JSON(
				writer,
				http.StatusOK,
				api.OrganizationResponse{
					Data: api.Organization{ID: value.ID, Name: value.Name},
				},
			)
		},
	)
}
