package config

import "github.com/zeroSal/went-web/controller"

type Iris struct {
	templateAutoReload bool
	registry           *controller.Registry
}

func NewIris(
	templateAutoReload bool,
	registry *controller.Registry,
) *Iris {
	return &Iris{
		templateAutoReload: templateAutoReload,
		registry:           registry,
	}
}

func (i *Iris) IsTemplateAutoReload() bool {
	return i.templateAutoReload
}

func (i *Iris) GetRegistry() *controller.Registry {
	return i.registry
}
