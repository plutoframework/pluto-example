package pluto

import (
	"context"
	"net/http"
	"strings"

	"github.com/plutoframework/pluto-example/internal/auth"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/features"
	"github.com/plutoframework/pluto-example/internal/paths"

	"github.com/go-chi/chi/v5"
	rbac "github.com/plutoframework/pluto-rbac"
)

type Registrar struct {
	router        chi.Router
	authenticator auth.Authenticator
	features      features.Evaluator
	policy        *rbac.Policy
}

func NewRegistrar(
	router chi.Router,
	authenticator auth.Authenticator,
	evaluator features.Evaluator,
	policy *rbac.Policy,
) *Registrar {
	return &Registrar{
		router:        router,
		authenticator: authenticator,
		features:      evaluator,
		policy:        policy,
	}
}

func (r *Registrar) Handle(endpoint paths.Endpoint, handler http.HandlerFunc) {
	r.router.MethodFunc(
		endpoint.Method,
		endpoint.Path,
		func(writer http.ResponseWriter, request *http.Request) {
			logger := Logger(request.Context()).With("endpoint_id", endpoint.ID)

			state, _ := request.Context().Value(stateKey).(*requestState)
			if state != nil {
				state.logger = logger
			}

			request = request.WithContext(
				context.WithValue(request.Context(), loggerKey, logger),
			)
			if endpoint.Authenticated {
				authorization := request.Header.Get("Authorization")

				scheme, token, found := strings.Cut(authorization, " ")
				if !found || !strings.EqualFold(scheme, "Bearer") ||
					strings.TrimSpace(token) != token {
					Error(writer, request, domain.ErrUnauthenticated)

					return
				}

				principal, err := r.authenticator.Authenticate(
					request.Context(),
					token,
					request.Header.Get("X-Organization-ID"),
				)
				if err != nil {
					Error(writer, request, err)

					return
				}

				if !principal.Verified() {
					Error(writer, request, domain.ErrUnauthenticated)

					return
				}

				logger = logger.With(
					"actor_id",
					principal.ActorID,
					"tenant_id",
					principal.TenantID,
				)
				if state != nil {
					state.logger = logger
				}

				ctx := context.WithValue(
					request.Context(),
					principalKey,
					principal,
				)

				request = request.WithContext(
					context.WithValue(ctx, loggerKey, logger),
				)
				if endpoint.Permission != "" &&
					!principal.Can(r.policy, endpoint.Permission) {
					Error(writer, request, domain.ErrForbidden)

					return
				}

				if endpoint.Feature != "" &&
					!r.features.Enabled(endpoint.Feature, principal) {
					Error(writer, request, domain.ErrFeatureDisabled)

					return
				}
			}

			handler(writer, request)
		},
	)
}
