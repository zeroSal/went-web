package auth

import "github.com/kataras/iris/v12"

type Interface interface {
	Authenticate(ctx iris.Context) (interface{}, bool)
	Middleware() iris.Handler
}
