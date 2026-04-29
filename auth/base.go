package auth

import (
	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/user"
)

type Base struct{}

func (b *Base) ExtractToken(ctx iris.Context) string {
	auth := ctx.GetHeader("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}

	return ""
}

func (b *Base) Middleware(
	authenticate func(ctx iris.Context) (user.Interface, bool),
) iris.Handler {
	return func(ctx iris.Context) {
		if _, ok := authenticate(ctx); !ok {
			ctx.StatusCode(iris.StatusUnauthorized)
			ctx.StopExecution()
			return
		}
		ctx.Next()
	}
}
