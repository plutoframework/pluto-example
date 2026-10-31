package app

import (
	"context"
	"errors"
	"net/http"

	"github.com/plutoframework/pluto-example/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	plutoauth "github.com/plutoframework/pluto-auth"
	"github.com/plutoframework/pluto-auth/httpauth"
	"github.com/plutoframework/pluto-auth/methods/oidc"
	"github.com/plutoframework/pluto-auth/methods/oidc/oidchttp"
	oidcpostgres "github.com/plutoframework/pluto-auth/stores/postgres/oidc"
)

func newOIDC(ctx context.Context, cfg config.Config,
	pool *pgxpool.Pool, manager *plutoauth.Manager,
	ingress func(http.Handler) http.Handler,
	clientHTTP *http.Client,
) (http.Handler, error) {
	if !cfg.OIDC.Enabled {
		return nil, nil
	}

	pending, err := oidcpostgres.New(pool, cfg.OIDC.EncryptionKey, 10000)
	if err != nil {
		return nil, errors.New("OIDC persistence configuration failed")
	}

	client, err := oidc.New(ctx, oidc.Config{
		Issuer:       cfg.OIDC.Issuer,
		ClientID:     cfg.OIDC.ClientID,
		ClientSecret: cfg.OIDC.ClientSecret,
		RedirectURL:  cfg.OIDC.CallbackURL,
		Transactions: pending, HTTPClient: clientHTTP,
	})
	if err != nil {
		return nil, errors.New("OIDC provider configuration failed")
	}

	service, err := oidc.NewService(
		client,
		pending,
		manager,
		cfg.OIDC.AllowEnrollment,
	)
	if err != nil {
		return nil, errors.New("OIDC authentication configuration failed")
	}

	return oidchttp.New(service, oidchttp.Config{
		Transport: httpauth.Config{
			TrustedOrigins: cfg.Auth.TrustedOrigins,
		},
		Ingress:         ingress,
		SuccessPath:     "/",
		CookieNamespace: cfg.OIDC.Issuer + ":" + cfg.OIDC.ClientID,
	})
}
