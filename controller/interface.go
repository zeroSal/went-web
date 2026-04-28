package controller

import "github.com/kataras/iris/v12"

type Base struct{}

type Configuration struct {
	Name   string
	Routes []Route
}

type Route struct {
	Path       string
	Method     string
	MethodName string
	Handler    iris.Handler
}

type Interface interface {
	GetConfiguration() Configuration
}
