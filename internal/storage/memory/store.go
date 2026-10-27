// Package memory is a local demonstration adapter; it is not durable storage.
package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/plutoframework/pluto-example/internal/domain"
)

type Store struct {
	mu          sync.RWMutex
	purchases   map[string]domain.Purchase
	memberships map[string]domain.Membership
	events      []domain.AuditEvent
}

func key(tenant, id string) string { return tenant + "\x00" + id }

func New(memberships []domain.Membership) *Store {
	store := &Store{
		purchases:   make(map[string]domain.Purchase),
		memberships: make(map[string]domain.Membership),
		events:      []domain.AuditEvent{},
	}
	for _, membership := range memberships {
		store.memberships[key(membership.TenantID, membership.ActorID)] = membership
	}

	return store
}

func (s *Store) Membership(
	_ context.Context,
	tenant, actor string,
) (domain.Membership, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	membership, ok := s.memberships[key(tenant, actor)]
	if !ok {
		return domain.Membership{}, domain.ErrNotFound
	}

	return membership, nil
}

func (s *Store) Memberships(
	_ context.Context,
	tenant string,
) ([]domain.Membership, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := []domain.Membership{}

	for _, membership := range s.memberships {
		if membership.TenantID == tenant {
			result = append(result, membership)
		}
	}

	sort.Slice(
		result,
		func(i, j int) bool { return result[i].ActorID < result[j].ActorID },
	)

	return result, nil
}

func (s *Store) ChangeRole(
	_ context.Context,
	tenant, actor string,
	role domain.Role,
	event domain.AuditEvent,
) error {
	if event.TenantID != tenant {
		return domain.ErrInvalid
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	membership, ok := s.memberships[key(tenant, actor)]
	if !ok {
		return domain.ErrNotFound
	}

	membership.Role = role
	s.memberships[key(tenant, actor)] = membership
	s.events = append(s.events, event)

	return nil
}

func (s *Store) Purchase(
	_ context.Context,
	tenant, id string,
) (domain.Purchase, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	purchase, ok := s.purchases[key(tenant, id)]
	if !ok {
		return domain.Purchase{}, domain.ErrNotFound
	}

	return purchase, nil
}

func (s *Store) Purchases(
	_ context.Context,
	tenant string,
) ([]domain.Purchase, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := []domain.Purchase{}

	for _, purchase := range s.purchases {
		if purchase.TenantID == tenant {
			result = append(result, purchase)
		}
	}

	sort.Slice(
		result,
		func(i, j int) bool { return result[i].ID < result[j].ID },
	)

	return result, nil
}

func (s *Store) SavePurchase(
	_ context.Context,
	purchase domain.Purchase,
	expected int64,
	event domain.AuditEvent,
) error {
	if event.TenantID != purchase.TenantID {
		return domain.ErrInvalid
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.purchases[key(purchase.TenantID, purchase.ID)]
	if (expected == 0 && exists) ||
		(expected != 0 && (!exists || existing.Version != expected)) {
		return domain.ErrConflict
	}

	s.purchases[key(purchase.TenantID, purchase.ID)] = purchase
	s.events = append(s.events, event)

	return nil
}

func (s *Store) Audit(
	_ context.Context,
	tenant, actor string,
) ([]domain.AuditEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := []domain.AuditEvent{}

	for _, event := range s.events {
		if event.TenantID == tenant && (actor == "" || event.ActorID == actor) {
			result = append(result, event)
		}
	}

	return result, nil
}
