package operations

import "context"

type SessionManager interface {
	Revoke(context.Context, string) error
}

type Sessions struct {
	Manager SessionManager
}

func (s Sessions) SignOut(ctx context.Context, token string) error {
	return s.Manager.Revoke(ctx, token)
}
