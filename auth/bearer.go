package auth

import (
	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/session"
	"github.com/zeroSal/went-web/user"
)

var _ Interface = (*Bearer)(nil)

type Bearer struct {
	Base
	SessionProvider session.ProviderInterface
}

func NewBearer() *Bearer {
	return &Bearer{}
}

func (b *Bearer) SetSessionProvider(provider session.ProviderInterface) {
	b.SessionProvider = provider
}

func (b *Bearer) Authenticate(ctx iris.Context) (user.Interface, bool) {
	token := b.ExtractToken(ctx)
	if token == "" {
		return nil, false
	}

	if b.SessionProvider == nil {
		return nil, false
	}

	u, err := b.SessionProvider.Load(token)
	if err != nil {
		return nil, false
	}

	return u, true
}

func (b *Bearer) Middleware() iris.Handler {
	return b.Base.Middleware(b.Authenticate)
}
