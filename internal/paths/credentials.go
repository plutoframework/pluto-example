package paths

import "github.com/plutoframework/pluto-example/internal/api"

// These routes are mounted directly through passwordhttp, before tenancy.
// They are never eligible for business response/idempotency caching.
var (
	Enroll = credentialEndpoint(
		"enroll",
		"Enroll account",
		"/auth/enroll",
		api.PasswordInput{},
		202,
	)
	RequestVerification = credentialEndpoint(
		"verification",
		"Resend verification",
		"/auth/verification",
		api.EmailInput{},
		202,
	)
	VerifyEmail = credentialEndpoint(
		"verify",
		"Verify email",
		"/auth/verify",
		api.ChallengeInput{},
		200,
	)
	SignIn = credentialEndpoint(
		"signin",
		"Sign in",
		"/auth/signin",
		api.PasswordInput{},
		200,
	)
	RequestRecovery = credentialEndpoint(
		"recovery",
		"Request recovery",
		"/auth/recovery",
		api.EmailInput{},
		202,
	)
	ResetPassword = credentialEndpoint(
		"reset",
		"Reset password",
		"/auth/reset",
		api.ResetInput{},
		200,
	)
	Logout = credentialEndpoint(
		"logout",
		"Sign out",
		"/auth/logout",
		nil,
		200,
	)
)

func credentialEndpoint(id, name, path string, input any, status int) Endpoint {
	return Endpoint{
		ID:            "credentials." + id,
		Name:          name,
		Description:   "POST only, exact trusted Origin required; opaque code response. No automatic business privileges. JSON is bounded to 8 KiB.",
		Method:        "POST",
		Path:          path,
		Authenticated: id == "logout",
		SessionOnly:   true,
		Request:       input,
		Response:      api.AuthResult{},
		SuccessStatus: status,
	}
}

func LocalCredentials() []Endpoint {
	return []Endpoint{
		Enroll,
		RequestVerification,
		VerifyEmail,
		SignIn,
		RequestRecovery,
		ResetPassword,
		Logout,
	}
}
