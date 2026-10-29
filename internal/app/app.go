// Package app is the explicit composition root for this example application.
package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/plutoframework/pluto-example/internal/auth"
	"github.com/plutoframework/pluto-example/internal/caching"
	"github.com/plutoframework/pluto-example/internal/config"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/features"
	"github.com/plutoframework/pluto-example/internal/operations"
	"github.com/plutoframework/pluto-example/internal/pluto"
	"github.com/plutoframework/pluto-example/internal/storage/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/plutoframework/pluto-auth/methods/password"
	authpostgres "github.com/plutoframework/pluto-auth/stores/postgres"
	idempostgres "github.com/plutoframework/pluto-core/stores/postgres"
	"github.com/plutoframework/pluto-core/telemetry"
)

type App struct {
	Accounts auth.AccountDisabler
	Handler  http.Handler
	Close    func()
}

type applicationServices struct {
	*operations.Service
	*operations.OrganizationService
	operations.AccessService
}

// New runs only the opt-in local demo. Production integrations are not inferred.
func New(
	ctx context.Context,
	settings config.Config,
	logger *slog.Logger,
) (*App, error) {
	return NewInstrumented(ctx, settings, logger, nil)
}

// NewInstrumented injects a runtime owned and shut down by the caller.
func NewInstrumented(ctx context.Context, settings config.Config,
	logger *slog.Logger, runtime *telemetry.Runtime,
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

	return newWithLocal(ctx, settings, logger, auth.LocalDependencies{
		Sender:  sender,
		Limiter: limiter,
		Ingress: ingress,
		DeliveryError: func(context.Context, error) {
			logger.Warn("authentication delivery failed; request a resend")
		},
	}, runtime)
}

// NewWithLocal makes delivery/limits explicit for tests and application wiring.
// Configuration still enforces the loopback development-only application scope.
func NewWithLocal(
	ctx context.Context,
	settings config.Config,
	logger *slog.Logger,
	dependencies auth.LocalDependencies,
) (*App, error) {
	return newWithLocal(ctx, settings, logger, dependencies, nil)
}

func newWithLocal(ctx context.Context, settings config.Config,
	logger *slog.Logger, dependencies auth.LocalDependencies,
	runtime *telemetry.Runtime,
) (*App, error) {
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	policy, err := domain.NewPolicy()
	if err != nil {
		return nil, err
	}

	var tracer pgx.QueryTracer
	if runtime != nil {
		tracer = runtime.PGXTracer()
	}

	store, err := postgres.OpenWithTracer(ctx, settings.Database.URL, tracer)
	if err != nil {
		return nil, errors.New("postgres connectivity failed")
	}

	ready := false

	var closeCache func()

	defer func() {
		if !ready {
			if closeCache != nil {
				closeCache()
			}

			store.Close()
		}
	}()

	dataCache, shutdownCache, err := caching.New(settings.Cache)
	if err != nil {
		return nil, err
	}

	closeCache = shutdownCache

	if runtime != nil {
		dataCache = runtime.Cache(dataCache)
	}

	environment := settings.Cache.Environment
	if environment == "" {
		environment = "development"
	}

	organizationService, err := operations.NewOrganizationService(
		store,
		dataCache,
		policy,
		environment,
	)
	if err != nil {
		return nil, err
	}

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
		Handler: RoutesWithTelemetry(
			settings.Server,
			logger,
			authentication,
			evaluator,
			applicationServices{
				Service:             service,
				OrganizationService: organizationService,
				AccessService:       operations.AccessService{Policy: policy},
			},
			policy,
			runtime,
		),
		Close: func() { closeCache(); store.Close() },
	}, nil
}
