package controller

import "github.com/kataras/iris/v12"

type Interface interface {
	Register(app *iris.Application)
}
