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
	"github.com/plutoframework/pluto-example/internal/pluto"
	"github.com/plutoframework/pluto-example/internal/storage/memory"
	"github.com/plutoframework/pluto-example/internal/storage/postgres"

	plutoauth "github.com/plutoframework/pluto-auth"
	"github.com/plutoframework/pluto-auth/httpauth"
	authmemory "github.com/plutoframework/pluto-auth/stores/memory"
	"github.com/plutoframework/pluto-core/idempotency"
	idemmemory "github.com/plutoframework/pluto-core/idempotency/memory"
	idempostgres "github.com/plutoframework/pluto-core/stores/postgres"
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

	var (
		idempotencyStore idempotency.Store
		repository       operations.Repository
		sessionStore     plutoauth.SessionStore
	)

	closeStore := func() {}

	switch settings.Database.Mode {
	case "memory":
		repository = memory.New(auth.DemoMemberships())

		idempotencyStore, err = idemmemory.New(1000)
		if err != nil {
			return nil, err
		}

		sessionStore, err = authmemory.New(1000)
		if err != nil {
			return nil, err
		}
	case "postgres":
		store, err := postgres.Open(ctx, settings.Database.URL)
		if err != nil {
			return nil, err
		}

		sessionStore, err = store.SessionStore()
		if err != nil {
			store.Close()

			return nil, err
		}

		idempotencyStore, err = idempostgres.New(store.Pool())
		if err != nil {
			store.Close()

			return nil, err
		}

		repository = store
		closeStore = store.Close
	}

	evaluator := features.NewSnapshot(settings.Features.Overrides())
	service := operations.New(repository, evaluator, policy)
	sessionConfig := plutoauth.DefaultConfig()
	sessionConfig.TTL = settings.Auth.SessionTTL
	sessionConfig.IdleTimeout = settings.Auth.IdleTimeout

	sessions, err := plutoauth.New(sessionStore, sessionConfig)
	if err != nil {
		closeStore()

		return nil, err
	}

	sessionHTTP, err := httpauth.New(sessions, httpauth.Config{
		TrustedOrigins: settings.Auth.TrustedOrigins,
		AllowBearer:    true,
		WriteError:     pluto.SessionError,
	})
	if err != nil {
		closeStore()

		return nil, err
	}

	idempotent, err := pluto.NewIdempotency(
		idempotencyStore,
		settings.Server.MaxBodyBytes,
	)
	if err != nil {
		closeStore()

		return nil, err
	}

	authentication := SessionAuthentication{
		Idempotency: idempotent,
		Resolver:    auth.Sessions{Memberships: repository},
		HTTP:        sessionHTTP,
		Operations: operations.Sessions{
			Verifier: auth.DemoCredentials{},
			Manager:  sessions,
		},
	}
	handler := Routes(
		settings.Server,
		logger,
		authentication,
		evaluator,
		service,
		policy,
	)

	return &App{
		Handler: handler,
		Close:   closeStore,
	}, nil
}
