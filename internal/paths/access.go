package paths

import (
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
)

var GetAccess = Endpoint{
	ID:            "access.get",
	Name:          "Read active organization capabilities",
	Description:   "Describe this request's verified role and effective capabilities for UI visibility. Business operations independently enforce authorization.",
	Method:        http.MethodGet,
	Path:          "/api/access",
	Authenticated: true,
	Response:      api.AccessResponse{},
	SuccessStatus: http.StatusOK,
}
