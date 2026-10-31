package app

import (
	"log/slog"
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/auth"
	"github.com/plutoframework/pluto-example/internal/config"
	"github.com/plutoframework/pluto-example/internal/features"
	accesshandlers "github.com/plutoframework/pluto-example/internal/handlers/access"
	audithandlers "github.com/plutoframework/pluto-example/internal/handlers/audit"
	featurehandlers "github.com/plutoframework/pluto-example/internal/handlers/features"
	"github.com/plutoframework/pluto-example/internal/handlers/memberships"
	"github.com/plutoframework/pluto-example/internal/handlers/organization"
	"github.com/plutoframework/pluto-example/internal/handlers/purchases"
	sessionhandlers "github.com/plutoframework/pluto-example/internal/handlers/sessions"
	"github.com/plutoframework/pluto-example/internal/paths"
	"github.com/plutoframework/pluto-example/internal/pluto"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/plutoframework/pluto-auth/httpauth"
	"github.com/plutoframework/pluto-core/httpidem"
	"github.com/plutoframework/pluto-core/telemetry"
	rbac "github.com/plutoframework/pluto-rbac"
)

type ApplicationOperations interface {
	accesshandlers.Operations
	organization.Operations
	purchases.Operations
	memberships.Operations
	audithandlers.Operations
}

type SessionAuthentication struct {
	OIDC        http.Handler
	Credentials http.Handler
	Idempotency *httpidem.Handler
	Resolver    auth.Authenticator
	HTTP        *httpauth.Transport
	Operations  sessionhandlers.Operations
}

func Routes(
	settings config.ServerConfig,
	logger *slog.Logger,
	authentication SessionAuthentication,
	evaluator features.Evaluator,
	operations ApplicationOperations,
	policy *rbac.Policy,
) http.Handler {
	return RoutesWithTelemetry(settings, logger, authentication, evaluator,
		operations, policy, nil)
}

// RoutesWithTelemetry installs instrumentation within Chi so route templates
// are available when the request finishes, without recording raw URL paths.
func RoutesWithTelemetry(settings config.ServerConfig, logger *slog.Logger,
	authentication SessionAuthentication, evaluator features.Evaluator,
	operations ApplicationOperations, policy *rbac.Policy,
	runtime *telemetry.Runtime,
) http.Handler {
	router := chi.NewRouter()
	if runtime != nil {
		router.Use(runtime.Middleware)
	}

	router.Use(pluto.SecurityHeaders)
	router.Use(pluto.Middleware(logger))

	if settings.Gzip {
		router.Use(middleware.Compress(5, "application/json"))
	}

	router.NotFound(func(writer http.ResponseWriter, request *http.Request) {
		pluto.JSON(writer, http.StatusNotFound, api.ErrorResponse{
			Error: api.Error{
				Code:      "not_found",
				Message:   "endpoint not found",
				RequestID: writer.Header().Get("X-Request-ID"),
			},
		})
	})

	router.MethodNotAllowed(
		func(writer http.ResponseWriter, request *http.Request) {
			pluto.JSON(writer, http.StatusMethodNotAllowed, api.ErrorResponse{
				Error: api.Error{
					Code:      "method_not_allowed",
					Message:   "method not allowed",
					RequestID: writer.Header().Get("X-Request-ID"),
				},
			})
		},
	)

	// Lifecycle routes run before tenant/membership and idempotency middleware.
	if authentication.Credentials != nil {
		for _, endpoint := range paths.LocalCredentials() {
			router.Method(endpoint.Method, endpoint.Path,
				http.StripPrefix("/auth", authentication.Credentials))
		}
	}

	if authentication.OIDC != nil {
		router.Mount(
			"/auth/oidc",
			http.StripPrefix("/auth/oidc", authentication.OIDC),
		)
	}

	router.Get(
		paths.OIDCProviders.Path,
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")

			providers := []string{}
			if authentication.OIDC != nil {
				providers = append(providers, "oidc")
			}

			pluto.JSON(w, 200, api.OIDCProviderList{Data: providers})
		},
	)

	registrar := pluto.NewRegistrar(
		router,
		authentication.Resolver,
		evaluator,
		policy,
		authentication.HTTP,
	)

	registrar.Idempotency = authentication.Idempotency
	if authentication.HTTP != nil && authentication.Operations != nil {
		sessionhandlers.Register(
			registrar,
			authentication.Operations,
			authentication.HTTP,
		)
	}

	purchases.Register(registrar, operations, settings.MaxBodyBytes)
	organization.Register(registrar, operations)
	accesshandlers.Register(registrar, operations)
	memberships.Register(registrar, operations, settings.MaxBodyBytes)
	featurehandlers.Register(registrar, evaluator)
	audithandlers.Register(registrar, operations)

	return router
}
