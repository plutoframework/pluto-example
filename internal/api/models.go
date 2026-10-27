// Package api owns explicit wire models shared by adapters and OpenAPI generation.
package api

import (
	"time"

	"github.com/plutoframework/pluto-example/internal/domain"
)

type CreatePurchaseRequest struct {
	Title       string `json:"title"`
	AmountCents int64  `json:"amount_cents"`
}

type DecisionRequest struct {
	Decision string `json:"decision"`
}

type ChangeRoleRequest struct {
	Role string `json:"role"`
}

type Purchase struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	CreatorID      string    `json:"creator_id"`
	Title          string    `json:"title"`
	AmountCents    int64     `json:"amount_cents"`
	Status         string    `json:"status"`
	Version        int64     `json:"version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Membership struct {
	ActorID     string `json:"actor_id"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type AuditEvent struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	ActorID        string    `json:"actor_id"`
	Action         string    `json:"action"`
	ResourceID     string    `json:"resource_id"`
	Outcome        string    `json:"outcome"`
	OccurredAt     time.Time `json:"occurred_at"`
}

type PurchaseResponse struct {
	Data Purchase `json:"data"`
}

type PurchasesResponse struct {
	Data []Purchase `json:"data"`
}

type MembershipsResponse struct {
	Data []Membership `json:"data"`
}

type AuditResponse struct {
	Data []AuditEvent `json:"data"`
}

type FeaturesResponse struct {
	Data map[string]bool `json:"data"`
}

type ChangedResponse struct {
	Data Changed `json:"data"`
}

type Changed struct {
	Updated bool `json:"updated"`
}

type ErrorResponse struct {
	Error Error `json:"error"`
}

type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func PurchaseModel(value domain.Purchase) Purchase {
	return Purchase{
		ID:             value.ID,
		OrganizationID: value.TenantID,
		CreatorID:      value.CreatorID,
		Title:          value.Title,
		AmountCents:    value.AmountCents,
		Status:         string(value.Status),
		Version:        value.Version,
		CreatedAt:      value.CreatedAt,
		UpdatedAt:      value.UpdatedAt,
	}
}

func PurchaseModels(values []domain.Purchase) []Purchase {
	result := make([]Purchase, 0, len(values))
	for _, value := range values {
		result = append(result, PurchaseModel(value))
	}

	return result
}
