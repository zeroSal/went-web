package auth

import (
	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/session"
	"github.com/zeroSal/went-web/user"
)

var _ Interface = (*Composite)(nil)

type Composite struct {
	Base
	authenticators  []Interface
	sessionProvider session.ProviderInterface
}

func NewComposite(authenticators ...Interface) *Composite {
	return &Composite{
		authenticators: authenticators,
	}
}

func (c *Composite) SetSessionProvider(provider session.ProviderInterface) {
	c.sessionProvider = provider
	for _, auth := range c.authenticators {
		auth.SetSessionProvider(provider)
	}
}

func (c *Composite) Authenticate(ctx iris.Context) (user.Interface, bool) {
	for _, auth := range c.authenticators {
		u, ok := auth.Authenticate(ctx)
		if ok {
			if u != nil {
				return u, true
			}

			continue
		}
	}
	return nil, false
}

func (b *Composite) Middleware() iris.Handler {
	return b.Base.Middleware(b.Authenticate)
}
