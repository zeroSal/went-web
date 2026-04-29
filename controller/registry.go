package controller

import (
	"reflect"

	"github.com/kataras/iris/v12"
)

type Handler struct {
	Controller string
	Method     string
	Handler    iris.Handler
}

type Registry struct {
	controllers []Interface
}

func NewRegistry() *Registry {
	return &Registry{
		controllers: []Interface{},
	}
}

func (r *Registry) Register(c Interface) {
	r.controllers = append(r.controllers, c)
}

func (r *Registry) Handlers() []Handler {
	var handlers []Handler
	for _, ctrl := range r.controllers {
		cfg := ctrl.GetConfiguration()

		ctrlVal := reflect.ValueOf(ctrl)
		ctrlType := reflect.TypeOf(ctrl)

		var methodType reflect.Type
		if ctrlType.Kind() == reflect.Ptr {
			methodType = ctrlType
		} else {
			methodType = reflect.PtrTo(ctrlType)
		}

		for i := 0; i < methodType.NumMethod(); i++ {
			m := methodType.Method(i)
			if m.Name == "GetConfiguration" {
				continue
			}

			if m.Type.NumIn() != 2 {
				continue
			}

			ctxType := reflect.TypeFor[iris.Context]()
			if m.Type.In(1).String() != ctxType.String() {
				continue
			}

			handlerVal := ctrlVal.MethodByName(m.Name)
			if !handlerVal.IsValid() {
				continue
			}

			handler, ok := handlerVal.Interface().(func(iris.Context))
			if !ok {
				continue
			}

			handlers = append(handlers, Handler{
				Controller: cfg.Name,
				Method:     m.Name,
				Handler:    handler,
			})
		}
	}
	return handlers
}
