package paths

import "github.com/plutoframework/pluto-example/internal/api"

var GetSession = Endpoint{
	ID:            "sessions.get",
	Name:          "Read current session",
	Description:   "Read verified session metadata without choosing an organization.",
	Method:        "GET",
	Path:          "/auth/session",
	Authenticated: true,
	SessionOnly:   true,
	Response:      api.SessionResponse{},
	SuccessStatus: 200,
}

var DeleteSession = Endpoint{
	ID:            "sessions.delete",
	Name:          "Sign out",
	Description:   "Revoke the current session before clearing its cookie. Does not require an active organization membership.",
	Method:        "DELETE",
	Path:          "/auth/session",
	Authenticated: true,
	SessionOnly:   true,
	Response:      api.ChangedResponse{},
	SuccessStatus: 200,
}
