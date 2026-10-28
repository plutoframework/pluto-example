// Package app is the explicit composition root for this example application.
package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/plutoframework/pluto-example/internal/auth"
	"github.com/plutoframework/pluto-example/internal/config"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/features"
	"github.com/plutoframework/pluto-example/internal/operations"
	"github.com/plutoframework/pluto-example/internal/pluto"
	"github.com/plutoframework/pluto-example/internal/storage/postgres"

	"github.com/plutoframework/pluto-auth/methods/password"
	authpostgres "github.com/plutoframework/pluto-auth/stores/postgres"
	idempostgres "github.com/plutoframework/pluto-core/stores/postgres"
)

type App struct {
	Accounts auth.AccountDisabler
	Handler  http.Handler
	Close    func()
}

// New runs only the opt-in local demo. Production integrations are not inferred.
func New(
	ctx context.Context,
	settings config.Config,
	logger *slog.Logger,
) (*App, error) {
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	sender, err := auth.NewDevelopmentSMTP(settings.Auth.DevelopmentSMTP)
	if err != nil {
		return nil, err
	}

	limiter, err := password.NewWindowLimiter(1000, 20, time.Minute, time.Now)
	if err != nil {
		return nil, err
	}

	ingress, err := auth.DevelopmentIngress()
	if err != nil {
		return nil, err
	}

	return NewWithLocal(ctx, settings, logger, auth.LocalDependencies{
		Sender:  sender,
		Limiter: limiter,
		Ingress: ingress,
		DeliveryError: func(context.Context, error) {
			logger.Warn("authentication delivery failed; request a resend")
		},
	})
}

// NewWithLocal makes delivery/limits explicit for tests and application wiring.
// Configuration still enforces the loopback development-only application scope.
func NewWithLocal(
	ctx context.Context,
	settings config.Config,
	logger *slog.Logger,
	dependencies auth.LocalDependencies,
) (*App, error) {
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	policy, err := domain.NewPolicy()
	if err != nil {
		return nil, err
	}

	store, err := postgres.Open(ctx, settings.Database.URL)
	if err != nil {
		return nil, errors.New("postgres connectivity failed")
	}

	ready := false
	defer func() {
		if !ready {
			store.Close()
		}
	}()

	localStore, err := authpostgres.NewLocal(store.Pool())
	if err != nil {
		return nil, err
	}

	credentials, credentialHTTP, err := auth.NewLocal(
		ctx,
		localStore,
		settings.Auth,
		dependencies,
	)
	if err != nil {
		return nil, err
	}

	idempotencyStore, err := idempostgres.New(store.Pool())
	if err != nil {
		return nil, err
	}

	idempotent, err := pluto.NewIdempotency(
		idempotencyStore,
		settings.Server.MaxBodyBytes,
	)
	if err != nil {
		return nil, err
	}

	evaluator := features.NewSnapshot(settings.Features.Overrides())
	service := operations.New(store, evaluator, policy)
	authentication := SessionAuthentication{
		Credentials: credentialHTTP,
		Idempotency: idempotent,
		Resolver:    auth.Sessions{Memberships: store},
		HTTP:        credentialHTTP.Transport(),
		Operations:  operations.Sessions{Manager: credentials.Sessions()},
	}
	ready = true

	return &App{
		Accounts: credentials,
		Handler: Routes(
			settings.Server,
			logger,
			authentication,
			evaluator,
			service,
			policy,
		),
		Close: store.Close,
	}, nil
}
