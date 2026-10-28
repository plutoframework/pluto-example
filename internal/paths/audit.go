package paths

import (
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/domain"
)

var ListAudit = Endpoint{
	Collection:    true,
	ID:            "audit.list",
	Name:          "Read audit history",
	Description:   "Members see their own actions; approvers and administrators see organization history.",
	Method:        http.MethodGet,
	Path:          "/api/audit",
	Authenticated: true,
	Permission:    domain.AuditRead,
	Response:      api.AuditResponse{},
	SuccessStatus: http.StatusOK,
}
