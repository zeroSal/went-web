package auth

import "github.com/kataras/iris/v12"

type Base struct{}

func (b *Base) ExtractToken(ctx iris.Context) string {
	auth := ctx.GetHeader("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}
