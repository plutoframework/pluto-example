package memberships

import (
	"context"

	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/paths"
	"github.com/plutoframework/pluto-example/internal/pluto"

	"github.com/plutoframework/pluto-core/pagination"
)

type Operations interface {
	MembershipsPage(
		context.Context,
		domain.Principal,
		pagination.Request,
	) (pagination.Page[domain.Membership], error)
	ChangeRole(context.Context, domain.Principal, string, domain.Role) error
}

type Handler struct {
	operations Operations
	bodyLimit  int64
}

func Register(
	registrar *pluto.Registrar,
	operations Operations,
	bodyLimit int64,
) {
	handler := &Handler{
		operations: operations,
		bodyLimit:  bodyLimit,
	}
	registrar.Handle(paths.ListMemberships, handler.list)
	registrar.Handle(paths.ChangeRole, handler.changeRole)
}
