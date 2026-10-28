package api

import (
	"time"

	plutoauth "github.com/plutoframework/pluto-auth"
)

type Session struct {
	ID        string    `json:"id"`
	ActorID   string    `json:"actor_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type SessionResponse struct {
	Data Session `json:"data"`
}

func SessionModel(session plutoauth.Session) Session {
	return Session{
		ID:        session.ID,
		ActorID:   session.SubjectID,
		ExpiresAt: session.ExpiresAt,
	}
}
