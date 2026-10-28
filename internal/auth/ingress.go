package auth

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/plutoframework/pluto-auth/methods/password"
)

// DevelopmentIngress uses the direct socket peer, ignoring forwarding headers.
// It has no queue and is suitable only for this loopback, single-process demo.
func DevelopmentIngress() (func(http.Handler) http.Handler, error) {
	clients, err := password.NewWindowLimiter(1000, 120, time.Minute, time.Now)
	if err != nil {
		return nil, err
	}

	active := make(chan struct{}, 4)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")

			host, _, err := net.SplitHostPort(r.RemoteAddr)

			peer := net.ParseIP(host)
			if err != nil || peer == nil || !peer.IsLoopback() {
				credentialError(w, r, http.StatusForbidden, "untrusted_client")

				return
			}

			if err := clients.Allow(r.Context(), peer.String()); err != nil {
				credentialError(
					w,
					r,
					http.StatusTooManyRequests,
					"request_limited",
				)

				return
			}

			select {
			case active <- struct{}{}:
				defer func() { <-active }()
			default:
				credentialError(
					w,
					r,
					http.StatusTooManyRequests,
					"request_limited",
				)

				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}, nil
}
