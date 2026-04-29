package controller

type Base struct{}

type Configuration struct {
	Name string
}

type Interface interface {
	GetConfiguration() Configuration
}
