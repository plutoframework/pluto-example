package purchases

import (
	"context"

	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/paths"
	"github.com/plutoframework/pluto-example/internal/pluto"
)

type Operations interface {
	Create(
		context.Context,
		domain.Principal,
		string,
		int64,
	) (domain.Purchase, error)
	List(context.Context, domain.Principal) ([]domain.Purchase, error)
	Get(context.Context, domain.Principal, string) (domain.Purchase, error)
	Submit(context.Context, domain.Principal, string) (domain.Purchase, error)
	Decide(
		context.Context,
		domain.Principal,
		string,
		domain.Status,
	) (domain.Purchase, error)
	Export(context.Context, domain.Principal) ([]domain.Purchase, error)
}

type Handler struct {
	operations Operations
	bodyLimit  int64
}

func Register(
	registrar *pluto.Registrar,
	operations Operations,
	bodyLimit int64,
) {
	handler := &Handler{
		operations: operations,
		bodyLimit:  bodyLimit,
	}
	registrar.Handle(paths.CreatePurchase, handler.create)
	registrar.Handle(paths.ListPurchases, handler.list)
	registrar.Handle(paths.ExportPurchases, handler.export)
	registrar.Handle(paths.GetPurchase, handler.get)
	registrar.Handle(paths.SubmitPurchase, handler.submit)
	registrar.Handle(paths.DecidePurchase, handler.decide)
}
