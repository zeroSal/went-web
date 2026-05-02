package controller

import (
	"github.com/kataras/iris/v12"
	"go.uber.org/fx"
)

type params struct {
	fx.In
	App         *iris.Application
	Controllers []Interface `group:"controllers"`
}

func Mount(p params) {
	for _, ctrl := range p.Controllers {
		ctrl.Register(p.App)
	}
}
