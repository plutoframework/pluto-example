package purchases

import (
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/paths"
	"github.com/plutoframework/pluto-example/internal/pluto"
)

func (h *Handler) create(writer http.ResponseWriter, request *http.Request) {
	var input api.CreatePurchaseRequest
	if err := pluto.DecodeJSON(
		writer,
		request,
		&input,
		h.bodyLimit,
	); err != nil {
		pluto.Error(writer, request, err)

		return
	}

	purchase, err := h.operations.Create(
		request.Context(),
		pluto.Principal(request),
		input.Title,
		input.AmountCents,
	)
	if err != nil {
		pluto.Error(writer, request, err)

		return
	}

	pluto.JSON(
		writer,
		paths.CreatePurchase.SuccessStatus,
		api.PurchaseResponse{Data: api.PurchaseModel(purchase)},
	)
}
