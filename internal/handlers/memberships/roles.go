package memberships

import (
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/paths"
	"github.com/plutoframework/pluto-example/internal/pluto"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) list(writer http.ResponseWriter, request *http.Request) {
	page, err := pluto.Pagination(request, paths.ListMemberships.ID)
	if err != nil {
		pluto.Error(writer, request, err)

		return
	}

	memberships, err := h.operations.MembershipsPage(
		request.Context(),
		pluto.Principal(request),
		page,
	)
	if err != nil {
		pluto.Error(writer, request, err)

		return
	}

	result := make([]api.Membership, 0, len(memberships.Data))
	for _, membership := range memberships.Data {
		result = append(result, api.Membership{
			ActorID:     membership.ActorID,
			DisplayName: membership.DisplayName,
			Role:        string(membership.Role),
		})
	}

	pluto.JSON(
		writer,
		paths.ListMemberships.SuccessStatus,
		api.MembershipsResponse{Data: result, Page: memberships.Page},
	)
}

func (h *Handler) changeRole(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var input api.ChangeRoleRequest
	if err := pluto.DecodeJSON(
		writer,
		request,
		&input,
		h.bodyLimit,
	); err != nil {
		pluto.Error(writer, request, err)

		return
	}

	err := h.operations.ChangeRole(
		request.Context(),
		pluto.Principal(request),
		chi.URLParam(request, "actorID"),
		domain.Role(input.Role),
	)
	if err != nil {
		pluto.Error(writer, request, err)

		return
	}

	pluto.JSON(
		writer,
		paths.ChangeRole.SuccessStatus,
		api.ChangedResponse{Data: api.Changed{Updated: true}},
	)
}
