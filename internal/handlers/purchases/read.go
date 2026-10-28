package purchases

import (
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/paths"
	"github.com/plutoframework/pluto-example/internal/pluto"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) get(writer http.ResponseWriter, request *http.Request) {
	purchase, err := h.operations.Get(
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
		paths.GetPurchase.SuccessStatus,
		api.PurchaseResponse{Data: api.PurchaseModel(purchase)},
	)
}

func (h *Handler) list(writer http.ResponseWriter, request *http.Request) {
	page, err := pluto.Pagination(request, paths.ListPurchases.ID)
	if err != nil {
		pluto.Error(writer, request, err)

		return
	}

	purchases, err := h.operations.ListPage(
		request.Context(),
		pluto.Principal(request),
		page,
	)
	if err != nil {
		pluto.Error(writer, request, err)

		return
	}

	pluto.JSON(
		writer,
		paths.ListPurchases.SuccessStatus,
		api.PurchasesResponse{
			Data: api.PurchaseModels(purchases.Data),
			Page: purchases.Page,
		},
	)
}

func (h *Handler) export(writer http.ResponseWriter, request *http.Request) {
	purchases, err := h.operations.Export(
		request.Context(),
		pluto.Principal(request),
	)
	if err != nil {
		pluto.Error(writer, request, err)

		return
	}

	writer.Header().
		Set("Content-Disposition", `attachment; filename="purchase-requests.json"`)
	pluto.JSON(
		writer,
		paths.ExportPurchases.SuccessStatus,
		api.PurchaseExportResponse{Data: api.PurchaseModels(purchases)},
	)
}
