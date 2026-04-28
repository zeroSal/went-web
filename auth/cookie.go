package auth

import "github.com/kataras/iris/v12"

type Cookie struct {
	Base
	CookieName string
}

func NewCookie(name string) *Cookie {
	return &Cookie{
		CookieName: name,
	}
}

func (c *Cookie) Authenticate(ctx iris.Context) (interface{}, bool) {
	token := ctx.GetCookie(c.CookieName)
	if token == "" {
		return nil, false
	}
	return token, true
}

func (c *Cookie) Middleware() iris.Handler {
	return func(ctx iris.Context) {
		if _, ok := c.Authenticate(ctx); !ok {
			ctx.StatusCode(iris.StatusUnauthorized)
			ctx.StopExecution()
			return
		}
		ctx.Next()
	}
}
