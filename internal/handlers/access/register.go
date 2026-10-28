package access

import (
	"context"
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/paths"
	"github.com/plutoframework/pluto-example/internal/pluto"
)

type Operations interface {
	Access(context.Context, domain.Principal) (domain.Access, error)
}

func Register(registrar *pluto.Registrar, operations Operations) {
	registrar.Handle(
		paths.GetAccess,
		func(w http.ResponseWriter, r *http.Request) {
			value, err := operations.Access(r.Context(), pluto.Principal(r))
			if err != nil {
				pluto.Error(w, r, err)

				return
			}

			permissions := make([]string, 0, len(value.Permissions))
			for _, p := range value.Permissions {
				permissions = append(permissions, string(p))
			}

			roles := make([]api.RoleOption, 0, len(value.Roles))
			for _, role := range value.Roles {
				roles = append(
					roles,
					api.RoleOption{
						ID:          string(role.ID),
						Name:        role.Name,
						Description: role.Description,
					},
				)
			}

			pluto.JSON(
				w,
				http.StatusOK,
				api.AccessResponse{
					Data: api.Access{
						ActorID:        value.ActorID,
						OrganizationID: value.OrganizationID,
						Role:           string(value.Role),
						RoleName:       value.RoleName,
						Permissions:    permissions,
						Roles:          roles,
					},
				},
			)
		},
	)
}
