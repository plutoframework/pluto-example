package operations

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/features"
)

type Service struct {
	repository Repository
	features   features.Evaluator
	now        func() time.Time
	newID      func() string
}

func New(repository Repository, evaluator features.Evaluator) *Service {
	return &Service{
		repository: repository,
		features:   evaluator,
		now:        time.Now,
		newID:      randomID,
	}
}

func randomID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic(err)
	}

	return hex.EncodeToString(value[:])
}

func require(principal domain.Principal, permission domain.Permission) error {
	if !principal.Verified() {
		return domain.ErrUnauthenticated
	}

	if !principal.Can(permission) {
		return domain.ErrForbidden
	}

	return nil
}

func (s *Service) event(
	principal domain.Principal,
	action, resource string,
) domain.AuditEvent {
	return domain.AuditEvent{
		ID:         s.newID(),
		TenantID:   principal.TenantID,
		ActorID:    principal.ActorID,
		Action:     action,
		ResourceID: resource,
		Outcome:    "success",
		OccurredAt: s.now().UTC(),
	}
}

func (s *Service) Create(
	ctx context.Context,
	principal domain.Principal,
	title string,
	cents int64,
) (domain.Purchase, error) {
	if err := require(principal, domain.PurchaseCreate); err != nil {
		return domain.Purchase{}, err
	}

	if err := domain.ValidatePurchase(title, cents); err != nil {
		return domain.Purchase{}, err
	}

	now := s.now().UTC()
	purchase := domain.Purchase{
		ID:          s.newID(),
		TenantID:    principal.TenantID,
		CreatorID:   principal.ActorID,
		Title:       strings.TrimSpace(title),
		AmountCents: cents,
		Status:      domain.Draft,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err := s.repository.SavePurchase(
		ctx,
		purchase,
		0,
		s.event(principal, "purchase.created", purchase.ID),
	)

	return purchase, err
}

func (s *Service) Get(
	ctx context.Context,
	principal domain.Principal,
	id string,
) (domain.Purchase, error) {
	if err := require(principal, domain.PurchaseRead); err != nil {
		return domain.Purchase{}, err
	}

	purchase, err := s.repository.Purchase(ctx, principal.TenantID, id)
	if err != nil {
		return domain.Purchase{}, err
	}

	if purchase.CreatorID != principal.ActorID &&
		!principal.Can(domain.PurchaseDecide) {
		return domain.Purchase{}, domain.ErrNotFound
	}

	return purchase, nil
}

func (s *Service) List(
	ctx context.Context,
	principal domain.Principal,
) ([]domain.Purchase, error) {
	if err := require(principal, domain.PurchaseRead); err != nil {
		return nil, err
	}

	purchases, err := s.repository.Purchases(ctx, principal.TenantID)
	if err != nil {
		return nil, err
	}

	visible := make([]domain.Purchase, 0, len(purchases))
	for _, purchase := range purchases {
		if purchase.CreatorID == principal.ActorID ||
			principal.Can(domain.PurchaseDecide) {
			visible = append(visible, purchase)
		}
	}

	return visible, nil
}

func (s *Service) Submit(
	ctx context.Context,
	principal domain.Principal,
	id string,
) (domain.Purchase, error) {
	if err := require(principal, domain.PurchaseCreate); err != nil {
		return domain.Purchase{}, err
	}

	purchase, err := s.repository.Purchase(ctx, principal.TenantID, id)
	if err != nil {
		return domain.Purchase{}, err
	}

	if purchase.CreatorID != principal.ActorID {
		return domain.Purchase{}, domain.ErrNotFound
	}

	if purchase.Status != domain.Draft {
		return domain.Purchase{}, domain.ErrConflict
	}

	return s.transition(
		ctx,
		principal,
		purchase,
		domain.Submitted,
		"purchase.submitted",
	)
}

func (s *Service) Decide(
	ctx context.Context,
	principal domain.Principal,
	id string,
	decision domain.Status,
) (domain.Purchase, error) {
	if err := require(principal, domain.PurchaseDecide); err != nil {
		return domain.Purchase{}, err
	}

	if decision != domain.Approved && decision != domain.Rejected {
		return domain.Purchase{}, domain.ErrInvalid
	}

	purchase, err := s.repository.Purchase(ctx, principal.TenantID, id)
	if err != nil {
		return domain.Purchase{}, err
	}

	if purchase.CreatorID == principal.ActorID {
		return domain.Purchase{}, domain.ErrForbidden
	}

	if purchase.Status != domain.Submitted {
		return domain.Purchase{}, domain.ErrConflict
	}

	return s.transition(
		ctx,
		principal,
		purchase,
		decision,
		"purchase."+string(decision),
	)
}

func (s *Service) transition(
	ctx context.Context,
	principal domain.Principal,
	purchase domain.Purchase,
	status domain.Status,
	action string,
) (domain.Purchase, error) {
	expectedVersion := purchase.Version
	purchase.Status = status
	purchase.Version++
	purchase.UpdatedAt = s.now().UTC()
	err := s.repository.SavePurchase(
		ctx,
		purchase,
		expectedVersion,
		s.event(principal, action, purchase.ID),
	)

	return purchase, err
}

func (s *Service) Export(
	ctx context.Context,
	principal domain.Principal,
) ([]domain.Purchase, error) {
	if err := require(principal, domain.PurchaseExport); err != nil {
		return nil, err
	}

	if !s.features.Enabled(features.PurchaseExport, principal) {
		return nil, domain.ErrFeatureDisabled
	}

	return s.List(ctx, principal)
}
