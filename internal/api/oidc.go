package api

type OIDCProviderList struct {
	Data []string `json:"data"`
}

type (
	OIDCAuthorizationData struct {
		AuthorizationURL string `json:"authorization_url"`
	}
	OIDCAuthorizationResponse struct {
		Data OIDCAuthorizationData `json:"data"`
	}
)
