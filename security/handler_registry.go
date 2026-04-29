package security

import "github.com/kataras/iris/v12"

type HandlerRegistry map[string]map[string]iris.Handler

func NewHandlerRegistry() HandlerRegistry {
	return make(HandlerRegistry)
}

func (r HandlerRegistry) Register(controller string, method string, handler iris.Handler) {
	if r[controller] == nil {
		r[controller] = make(map[string]iris.Handler)
	}
	r[controller][method] = handler
}

func (r HandlerRegistry) Get(controller, method string) iris.Handler {
	if r[controller] != nil {
		return r[controller][method]
	}
	return nil
}

func (r HandlerRegistry) Range(fn func(controller, method string, handler iris.Handler)) {
	for controller, methods := range r {
		for method, handler := range methods {
			fn(controller, method, handler)
		}
	}
}
