package paths

import "github.com/plutoframework/pluto-example/internal/api"

// OIDC routes bypass business idempotency and tenant selection. The native
// handler validates browser intent and sets Secure/HttpOnly cookies.
var (
	OIDCBegin = Endpoint{
		ID:            "oidc.begin",
		Name:          "Begin OIDC sign-in",
		Description:   "Trusted-Origin POST; redirects to configured provider with S256 PKCE.",
		Method:        "POST",
		Path:          "/auth/oidc/begin",
		SessionOnly:   true,
		SuccessStatus: 303,
		Response:      api.OIDCAuthorizationResponse{},
	}
	OIDCCallback = Endpoint{
		ID:            "oidc.callback",
		Name:          "Complete OIDC sign-in",
		Description:   "One-use browser/state/nonce callback; no automatic business grants.",
		Method:        "GET",
		Path:          "/auth/oidc/callback",
		SessionOnly:   true,
		SuccessStatus: 303,
	}
	OIDCProviders = Endpoint{
		ID:            "oidc.providers",
		Name:          "Configured sign-in providers",
		Description:   "Public provider IDs only; no secrets or endpoint configuration.",
		Method:        "GET",
		Path:          "/auth/providers",
		SuccessStatus: 200,
		Response:      api.OIDCProviderList{},
	}
)
