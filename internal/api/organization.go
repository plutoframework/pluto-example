package api

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type OrganizationResponse struct {
	Data Organization `json:"data"`
}
