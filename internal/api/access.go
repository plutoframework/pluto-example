package api

type RoleOption struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Access struct {
	ActorID        string       `json:"actor_id"`
	OrganizationID string       `json:"organization_id"`
	Role           string       `json:"role"`
	RoleName       string       `json:"role_name"`
	Permissions    []string     `json:"permissions"`
	Roles          []RoleOption `json:"roles"`
}

type AccessResponse struct {
	Data Access `json:"data"`
}
