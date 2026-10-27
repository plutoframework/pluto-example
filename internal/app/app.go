// Package app is the explicit composition root for this example application.
package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/plutoframework/pluto-example/internal/auth"
	"github.com/plutoframework/pluto-example/internal/config"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/features"
	"github.com/plutoframework/pluto-example/internal/operations"
	"github.com/plutoframework/pluto-example/internal/storage/memory"
	"github.com/plutoframework/pluto-example/internal/storage/postgres"
)

type App struct {
	Handler http.Handler
	Close   func()
}

func New(
	ctx context.Context,
	settings config.Config,
	logger *slog.Logger,
) (*App, error) {
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	policy, err := domain.NewPolicy()
	if err != nil {
		return nil, err
	}

	var repository operations.Repository

	closeStore := func() {}

	switch settings.Database.Mode {
	case "memory":
		repository = memory.New(auth.DemoMemberships())
	case "postgres":
		store, err := postgres.Open(ctx, settings.Database.URL)
		if err != nil {
			return nil, err
		}

		repository = store
		closeStore = store.Close
	}

	evaluator := features.NewSnapshot(settings.Features.Overrides())
	service := operations.New(repository, evaluator, policy)
	authenticator := auth.NewDemo(repository)
	handler := Routes(
		settings.Server,
		logger,
		authenticator,
		evaluator,
		service,
		policy,
	)

	return &App{
		Handler: handler,
		Close:   closeStore,
	}, nil
}
