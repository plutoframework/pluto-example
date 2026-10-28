package postgres

import (
	plutoauth "github.com/plutoframework/pluto-auth"
	authpostgres "github.com/plutoframework/pluto-auth/stores/postgres"
)

func (s *Store) SessionStore() (plutoauth.SessionStore, error) {
	return authpostgres.New(s.pool)
}
