package operations

import (
	"context"

	"github.com/plutoframework/pluto-example/internal/domain"

	plutoauth "github.com/plutoframework/pluto-auth"
)

type SubjectVerifier interface {
	Subject(context.Context, string) (string, error)
}

type SessionManager interface {
	Start(context.Context, string) (plutoauth.IssuedSession, error)
	Revoke(context.Context, string) error
}

type Sessions struct {
	Verifier SubjectVerifier
	Manager  SessionManager
}

func (s Sessions) SignIn(
	ctx context.Context,
	credential string,
) (plutoauth.IssuedSession, error) {
	subject, err := s.Verifier.Subject(ctx, credential)
	if err != nil {
		return plutoauth.IssuedSession{}, err
	}

	if subject == "" {
		return plutoauth.IssuedSession{}, domain.ErrUnauthenticated
	}

	return s.Manager.Start(ctx, subject)
}

func (s Sessions) SignOut(ctx context.Context, token string) error {
	return s.Manager.Revoke(ctx, token)
}
