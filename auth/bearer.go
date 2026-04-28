package auth

import "github.com/kataras/iris/v12"

type Bearer struct {
	Base
}

func NewBearer() *Bearer {
	return &Bearer{}
}

func (b *Bearer) Authenticate(ctx iris.Context) (interface{}, bool) {
	token := b.ExtractToken(ctx)
	if token == "" {
		return nil, false
	}
	return token, true
}

func (b *Bearer) Middleware() iris.Handler {
	return func(ctx iris.Context) {
		if _, ok := b.Authenticate(ctx); !ok {
			ctx.StatusCode(iris.StatusUnauthorized)
			ctx.StopExecution()
			return
		}
		ctx.Next()
	}
}
