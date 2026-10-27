package purchases

import (
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/paths"
	"github.com/plutoframework/pluto-example/internal/pluto"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) submit(writer http.ResponseWriter, request *http.Request) {
	purchase, err := h.operations.Submit(
		request.Context(),
		pluto.Principal(request),
		chi.URLParam(request, "purchaseID"),
	)
	if err != nil {
		pluto.Error(writer, request, err)

		return
	}

	pluto.JSON(
		writer,
		paths.SubmitPurchase.SuccessStatus,
		api.PurchaseResponse{Data: api.PurchaseModel(purchase)},
	)
}

func (h *Handler) decide(writer http.ResponseWriter, request *http.Request) {
	var input api.DecisionRequest
	if err := pluto.DecodeJSON(
		writer,
		request,
		&input,
		h.bodyLimit,
	); err != nil {
		pluto.Error(writer, request, err)

		return
	}

	purchase, err := h.operations.Decide(
		request.Context(),
		pluto.Principal(request),
		chi.URLParam(request, "purchaseID"),
		domain.Status(input.Decision),
	)
	if err != nil {
		pluto.Error(writer, request, err)

		return
	}

	pluto.JSON(
		writer,
		paths.DecidePurchase.SuccessStatus,
		api.PurchaseResponse{Data: api.PurchaseModel(purchase)},
	)
}
