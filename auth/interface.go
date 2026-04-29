package auth

import (
	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/session"
	"github.com/zeroSal/went-web/user"
)

type Interface interface {
	Authenticate(ctx iris.Context) (user.Interface, bool)
	Middleware() iris.Handler
	SetSessionProvider(provider session.ProviderInterface)
}
