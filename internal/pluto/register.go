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
	"github.com/plutoframework/pluto-auth/httpauth"
	"github.com/plutoframework/pluto-core/httpidem"
	"github.com/plutoframework/pluto-core/idempotency"
	rbac "github.com/plutoframework/pluto-rbac"
)

type Registrar struct {
	Idempotency   *httpidem.Handler
	router        chi.Router
	authenticator auth.Authenticator
	features      features.Evaluator
	policy        *rbac.Policy
	sessions      *httpauth.Transport
}

func NewRegistrar(
	router chi.Router,
	authenticator auth.Authenticator,
	evaluator features.Evaluator,
	policy *rbac.Policy,
	sessions *httpauth.Transport,
) *Registrar {
	return &Registrar{
		router:        router,
		authenticator: authenticator,
		features:      evaluator,
		policy:        policy,
		sessions:      sessions,
	}
}

func (r *Registrar) Handle(endpoint paths.Endpoint, handler http.HandlerFunc) {
	invoke := http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			logger := Logger(request.Context()).With("endpoint_id", endpoint.ID)

			state, _ := request.Context().Value(stateKey).(*requestState)
			if state != nil {
				state.logger = logger
			}

			request = request.WithContext(
				context.WithValue(request.Context(), loggerKey, logger),
			)
			if endpoint.Authenticated && !endpoint.SessionOnly {
				authorization := request.Header.Get("Authorization")

				scheme, token, found := strings.Cut(authorization, " ")

				_, sessionVerified := httpauth.FromContext(request.Context())
				if !sessionVerified &&
					(!found || !strings.EqualFold(scheme, "Bearer") ||
						strings.TrimSpace(token) != token) {
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

			if endpoint.Idempotent {
				if r.Idempotency == nil {
					JSON(
						writer,
						503,
						map[string]any{
							"error": map[string]any{
								"code":    "idempotency_unavailable",
								"message": "Service Unavailable",
								"request_id": writer.Header().
									Get("X-Request-ID"),
							},
						},
					)

					return
				}

				p := Principal(request)
				r.Idempotency.Serve(
					writer,
					request,
					idempotency.Scope{
						Subject:   p.ActorID,
						Tenant:    p.TenantID,
						Operation: endpoint.ID,
					},
					handler,
				)

				return
			}

			handler(writer, request)
		},
	)

	var secured http.Handler = invoke
	if endpoint.Authenticated && r.sessions != nil {
		secured = r.sessions.Require(invoke)
	} else if endpoint.SessionOnly {
		secured = http.HandlerFunc(
			func(w http.ResponseWriter, request *http.Request) {
				Error(w, request, domain.ErrUnauthenticated)
			},
		)
	}

	r.router.Method(endpoint.Method, endpoint.Path, secured)
}
