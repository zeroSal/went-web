package user

type Interface interface {
	GetID() any
	GetUsername() string
	GetRoles() []string
	HasRole(role string) bool
}
