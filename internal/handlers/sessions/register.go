package sessions

import (
	"context"
	"net/http"

	"github.com/plutoframework/pluto-example/internal/api"
	"github.com/plutoframework/pluto-example/internal/domain"
	"github.com/plutoframework/pluto-example/internal/paths"
	"github.com/plutoframework/pluto-example/internal/pluto"

	"github.com/plutoframework/pluto-auth/httpauth"
)

type Operations interface {
	SignOut(context.Context, string) error
}

func Register(
	registrar *pluto.Registrar,
	operations Operations,
	transport *httpauth.Transport,
) {
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
