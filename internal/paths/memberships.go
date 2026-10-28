package paths

import (
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/domain"
)

var ListMemberships = Endpoint{
	Collection:    true,
	ID:            "memberships.list",
	Name:          "List organization members",
	Description:   "Administrators list memberships in the verified organization.",
	Method:        http.MethodGet,
	Path:          "/api/memberships",
	Authenticated: true,
	Permission:    domain.MembershipManage,
	Response:      api.MembershipsResponse{},
	SuccessStatus: http.StatusOK,
}

var ChangeRole = Endpoint{
	Idempotent:    true,
	ID:            "memberships.change_role",
	Name:          "Change a member role",
	Description:   "Administrators change another existing member's role. Self changes are forbidden.",
	Method:        http.MethodPatch,
	Path:          "/api/memberships/{actorID}",
	Authenticated: true,
	Permission:    domain.MembershipManage,
	Request:       api.ChangeRoleRequest{},
	Response:      api.ChangedResponse{},
	SuccessStatus: http.StatusOK,
}
