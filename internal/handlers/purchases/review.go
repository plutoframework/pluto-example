package purchases

import (
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/pluto"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) review(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	result, err := h.operations.ReviewPurchase(r.Context(), pluto.Principal(r),
		chi.URLParam(r, "purchaseID"))
	if err != nil {
		pluto.Error(w, r, err)

		return
	}

	pluto.JSON(w, http.StatusOK, api.PurchaseReviewResponse{
		Data: api.PurchaseReview{
			Summary:   result.Summary,
			NextSteps: result.NextSteps, Demo: result.Demo,
		},
	})
}
