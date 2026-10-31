package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/plutoframework/pluto-example/internal/config"

	plutoauth "github.com/plutoframework/pluto-auth"
	"github.com/plutoframework/pluto-auth/httpauth"
	"github.com/plutoframework/pluto-auth/methods/password"
	"github.com/plutoframework/pluto-auth/methods/password/passwordhttp"
)

// AccountDisabler is a trusted administration seam, never a public HTTP route.
type AccountDisabler interface {
	Disable(context.Context, string) error
}

// LocalDependencies are supplied by the application, never request input.
// Production needs a real sender, shared limits and trusted-client ingress.
type LocalDependencies struct {
	// OIDCHTTPClient is an optional host-owned transport; nil uses bounded defaults.
	OIDCHTTPClient *http.Client
	Sender         password.Sender
	Limiter        password.AttemptLimiter
	Ingress        func(http.Handler) http.Handler
	Policy         password.PasswordPolicy
	DeliveryError  func(context.Context, error)
}

func NewLocal(ctx context.Context, store password.AccountStore,
	settings config.AuthConfig, dependencies LocalDependencies,
) (*password.Service, *passwordhttp.Handler, error) {
	hasher, err := password.NewHasher(password.DefaultConfig())
	if err != nil {
		return nil, nil, err
	}

	sessions := plutoauth.DefaultConfig()
	sessions.TTL = settings.SessionTTL
	sessions.IdleTimeout = settings.IdleTimeout
	lifecycle := password.DefaultLifecycleConfig()
	lifecycle.Policy = dependencies.Policy
	lifecycle.OnDeliveryError = dependencies.DeliveryError

	service, err := password.NewService(ctx, store, hasher,
		dependencies.Limiter, dependencies.Sender, sessions, lifecycle)
	if err != nil {
		return nil, nil, err
	}

	handler, err := passwordhttp.New(service, passwordhttp.Config{
		Transport: httpauth.Config{
			TrustedOrigins: settings.TrustedOrigins,
			AllowBearer:    true,
			WriteError:     credentialError,
		},
		Ingress: dependencies.Ingress,
	})
	if err != nil {
		return nil, nil, err
	}

	return service, handler, nil
}

func credentialError(
	w http.ResponseWriter,
	_ *http.Request,
	status int,
	code string,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(map[string]string{"code": code})
}
