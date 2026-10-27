package features

import (
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/features"
	"github.com/plutoframework/pluto-example/internal/paths"
	"github.com/plutoframework/pluto-example/internal/pluto"
)

func Register(registrar *pluto.Registrar, evaluator features.Evaluator) {
	registrar.Handle(
		paths.GetFeatures,
		func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Cache-Control", "no-store")

			values := features.Visible(evaluator, pluto.Principal(request))
			pluto.JSON(
				writer,
				paths.GetFeatures.SuccessStatus,
				api.FeaturesResponse{Data: values},
			)
		},
	)
}
