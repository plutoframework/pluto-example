package operations

import (
	"context"

	"github.com/plutoframework/pluto-example/internal/domain"

	"github.com/plutoframework/pluto-core/pagination"
)

func (s *Service) ListPage(
	ctx context.Context,
	p domain.Principal,
	request pagination.Request,
) (pagination.Page[domain.Purchase], error) {
	if err := s.require(p, domain.PurchaseRead); err != nil {
		return pagination.Page[domain.Purchase]{}, err
	}

	if err := request.Validate(); err != nil {
		return pagination.Page[domain.Purchase]{}, err
	}

	actor := p.ActorID
	if p.Can(s.policy, domain.PurchaseDecide) {
		actor = ""
	}

	rows, err := s.repository.PurchasePage(ctx, p.TenantID, actor, request)
	if err != nil {
		return pagination.Page[domain.Purchase]{}, err
	}

	return pagination.Finish(
		request,
		rows,
		func(v domain.Purchase) pagination.Key { return pagination.Key{ID: v.ID} },
	)
}

func (s *Service) MembershipsPage(
	ctx context.Context,
	p domain.Principal,
	request pagination.Request,
) (pagination.Page[domain.Membership], error) {
	if err := s.require(p, domain.MembershipManage); err != nil {
		return pagination.Page[domain.Membership]{}, err
	}

	if err := request.Validate(); err != nil {
		return pagination.Page[domain.Membership]{}, err
	}

	rows, err := s.repository.MembershipPage(ctx, p.TenantID, request)
	if err != nil {
		return pagination.Page[domain.Membership]{}, err
	}

	return pagination.Finish(
		request,
		rows,
		func(v domain.Membership) pagination.Key { return pagination.Key{ID: v.ActorID} },
	)
}

func (s *Service) AuditPage(
	ctx context.Context,
	p domain.Principal,
	request pagination.Request,
) (pagination.Page[domain.AuditEvent], error) {
	if err := s.require(p, domain.AuditRead); err != nil {
		return pagination.Page[domain.AuditEvent]{}, err
	}

	if err := request.Validate(); err != nil {
		return pagination.Page[domain.AuditEvent]{}, err
	}

	actor := p.ActorID
	if p.Can(s.policy, domain.AuditReadAll) {
		actor = ""
	}

	rows, err := s.repository.AuditPage(ctx, p.TenantID, actor, request)
	if err != nil {
		return pagination.Page[domain.AuditEvent]{}, err
	}

	return pagination.Finish(
		request,
		rows,
		func(v domain.AuditEvent) pagination.Key { return domain.AuditKey(v) },
	)
}
