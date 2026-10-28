package sessions

import (
	"context"
	"net/http"
	"strings"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/paths"
	"github.com/plutoframework/pluto-example/internal/pluto"

	plutoauth "github.com/plutoframework/pluto-auth"
	"github.com/plutoframework/pluto-auth/httpauth"
)

type Operations interface {
	SignIn(context.Context, string) (plutoauth.IssuedSession, error)
	SignOut(context.Context, string) error
}

func Register(
	registrar *pluto.Registrar,
	operations Operations,
	transport *httpauth.Transport,
) {
	create := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := r.Header.Values("Authorization")
		if len(headers) != 1 {
			pluto.Error(w, r, domain.ErrUnauthenticated)

			return
		}

		scheme, credential, ok := strings.Cut(headers[0], " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") ||
			strings.TrimSpace(credential) != credential {
			pluto.Error(w, r, domain.ErrUnauthenticated)

			return
		}

		issued, err := operations.SignIn(r.Context(), credential)
		if err != nil {
			pluto.Error(w, r, err)

			return
		}

		if err := transport.WriteCookie(w, issued); err != nil {
			pluto.Error(w, r, err)

			return
		}

		pluto.JSON(
			w,
			paths.CreateSession.SuccessStatus,
			api.SessionResponse{Data: api.SessionModel(issued.Session)},
		)
	})
	registrar.Handle(paths.CreateSession, transport.Guard(create).ServeHTTP)
	registrar.Handle(
		paths.GetSession,
		func(w http.ResponseWriter, r *http.Request) {
			session, ok := httpauth.FromContext(r.Context())
			if !ok {
				pluto.Error(w, r, domain.ErrUnauthenticated)

				return
			}

			pluto.JSON(
				w,
				paths.GetSession.SuccessStatus,
				api.SessionResponse{Data: api.SessionModel(session)},
			)
		},
	)
	registrar.Handle(
		paths.DeleteSession,
		func(w http.ResponseWriter, r *http.Request) {
			token, err := transport.Token(r)
			if err != nil {
				pluto.Error(w, r, domain.ErrUnauthenticated)

				return
			}

			if err := operations.SignOut(r.Context(), token); err != nil {
				pluto.Error(w, r, err)

				return
			}

			transport.ClearCookie(w)
			pluto.JSON(
				w,
				paths.DeleteSession.SuccessStatus,
				api.ChangedResponse{Data: api.Changed{Updated: true}},
			)
		},
	)
}
