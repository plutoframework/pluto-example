package paths

import (
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/domain"
)

var GetOrganization = Endpoint{
	ID:            "organization.get",
	Name:          "Read organization profile",
	Description:   "Read cached display metadata after live session, membership and permission checks.",
	Method:        http.MethodGet,
	Path:          "/api/organization",
	Authenticated: true,
	Permission:    domain.OrganizationRead,
	Response:      api.OrganizationResponse{},
	SuccessStatus: http.StatusOK,
}
