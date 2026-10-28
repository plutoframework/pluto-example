package paths

import (
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/features"
)

var CreatePurchase = Endpoint{
	Idempotent:    true,
	ID:            "purchases.create",
	Name:          "Create purchase request",
	Description:   "Create a draft request in the verified active organization.",
	Method:        http.MethodPost,
	Path:          "/api/purchases",
	Authenticated: true,
	Permission:    domain.PurchaseCreate,
	Request:       api.CreatePurchaseRequest{},
	Response:      api.PurchaseResponse{},
	SuccessStatus: http.StatusCreated,
}

var ListPurchases = Endpoint{
	Collection:    true,
	ID:            "purchases.list",
	Name:          "List visible purchase requests",
	Description:   "Members see their own requests; approvers and administrators see all requests in their organization.",
	Method:        http.MethodGet,
	Path:          "/api/purchases",
	Authenticated: true,
	Permission:    domain.PurchaseRead,
	Response:      api.PurchasesResponse{},
	SuccessStatus: http.StatusOK,
}

var GetPurchase = Endpoint{
	ID:            "purchases.get",
	Name:          "Read purchase request",
	Description:   "Read an owned request or an organization request visible to an approver.",
	Method:        http.MethodGet,
	Path:          "/api/purchases/{purchaseID}",
	Authenticated: true,
	Permission:    domain.PurchaseRead,
	Response:      api.PurchaseResponse{},
	SuccessStatus: http.StatusOK,
}

var SubmitPurchase = Endpoint{
	Idempotent:    true,
	ID:            "purchases.submit",
	Name:          "Submit purchase request",
	Description:   "The creator submits a draft for approval. Other transitions return conflict.",
	Method:        http.MethodPost,
	Path:          "/api/purchases/{purchaseID}/submit",
	Authenticated: true,
	Permission:    domain.PurchaseCreate,
	Response:      api.PurchaseResponse{},
	SuccessStatus: http.StatusOK,
}

var DecidePurchase = Endpoint{
	Idempotent:    true,
	ID:            "purchases.decide",
	Name:          "Decide purchase request",
	Description:   "An approver approves or rejects a submitted request created by someone else.",
	Method:        http.MethodPost,
	Path:          "/api/purchases/{purchaseID}/decision",
	Authenticated: true,
	Permission:    domain.PurchaseDecide,
	Request:       api.DecisionRequest{},
	Response:      api.PurchaseResponse{},
	SuccessStatus: http.StatusOK,
}

var ExportPurchases = Endpoint{
	ID:            "purchases.export",
	Name:          "Export purchase requests",
	Description:   "Return a JSON export for an authorized approver when purchase_export is enabled.",
	Method:        http.MethodGet,
	Path:          "/api/purchases/export",
	Authenticated: true,
	Permission:    domain.PurchaseExport,
	Feature:       features.PurchaseExport,
	Response:      api.PurchaseExportResponse{},
	SuccessStatus: http.StatusOK,
}
