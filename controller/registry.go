package controller

type Registry struct {
	constructors []any
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(constructor any) {
	r.constructors = append(r.constructors, constructor)
}

func (r *Registry) All() []any {
	return r.constructors
}
