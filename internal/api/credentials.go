package api

// AuthResult deliberately contains no account ID, token or challenge state.
type AuthResult struct {
	Code string `json:"code"`
}

type PasswordInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type EmailInput struct {
	Email string `json:"email"`
}

type ChallengeInput struct {
	Token string `json:"token"`
}

type ResetInput struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}
