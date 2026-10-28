package domain

type Access struct {
	ActorID        string
	OrganizationID string
	Role           Role
	RoleName       string
	Permissions    []Permission
	Roles          []RoleDefinition
}
