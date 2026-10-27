package paths

import (
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
)

var GetFeatures = Endpoint{
	ID:            "features.get",
	Name:          "Read visible feature flags",
	Description:   "Return allowlisted effective flags for the authenticated actor and verified organization; flags do not grant permissions.",
	Method:        http.MethodGet,
	Path:          "/api/features",
	Authenticated: true,
	Response:      api.FeaturesResponse{},
	SuccessStatus: http.StatusOK,
}
