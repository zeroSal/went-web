package auth

import (
	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/session"
	"github.com/zeroSal/went-web/user"
)

type Cookie struct {
	Base
	CookieName      string
	SessionProvider session.ProviderInterface
}

func NewCookie(name string) *Cookie {
	return &Cookie{
		CookieName: name,
	}
}

func (c *Cookie) SetSessionProvider(provider session.ProviderInterface) {
	c.SessionProvider = provider
}

func (c *Cookie) Authenticate(ctx iris.Context) (user.Interface, bool) {
	token := ctx.GetCookie(c.CookieName)
	if token == "" {
		return nil, false
	}

	if c.SessionProvider == nil {
		return nil, false
	}

	u, err := c.SessionProvider.Load(token)
	if err != nil {
		return nil, false
	}

	return u, true
}

func (b *Cookie) Middleware() iris.Handler {
	return b.Base.Middleware(b.Authenticate)
}
