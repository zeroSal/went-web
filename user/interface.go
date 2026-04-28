package user

import "github.com/kataras/iris/v12"

type Interface interface {
	GetID() any
	GetUsername() string
	GetRoles() []string
	HasRole(role string) bool
}

type Provider interface {
	LoadByUsername(username string) (Interface, error)
	LoadByID(id any) (Interface, error)
}

type RoleChecker interface {
	CheckRole(ctx iris.Context, user Interface, role string) bool
}

type RoleCheckerFunc func(ctx iris.Context, u Interface, role string) bool

func (f RoleCheckerFunc) CheckRole(ctx iris.Context, u Interface, role string) bool {
	return f(ctx, u, role)
}
