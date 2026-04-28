package auth

import "github.com/kataras/iris/v12"

type Composite struct {
	authenticators []Interface
}

func NewComposite(authenticators ...Interface) *Composite {
	return &Composite{
		authenticators: authenticators,
	}
}

func (c *Composite) Authenticate(ctx iris.Context) (interface{}, bool) {
	for _, auth := range c.authenticators {
		if user, ok := auth.Authenticate(ctx); ok {
			return user, true
		}
	}
	return nil, false
}

func (c *Composite) Middleware() iris.Handler {
	return func(ctx iris.Context) {
		if _, ok := c.Authenticate(ctx); !ok {
			ctx.StatusCode(iris.StatusUnauthorized)
			ctx.StopExecution()
			return
		}
		ctx.Next()
	}
}
