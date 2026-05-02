package wentweb

import (
	"github.com/kataras/iris/v12"
	"go.uber.org/fx"
)

var Bundle = fx.Options(
	fx.Provide(newIris),
)

func newIris() *iris.Application {
	return iris.New()
}
