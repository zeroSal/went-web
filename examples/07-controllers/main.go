package main

import (
	"log"

	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/controller"
	"github.com/zeroSal/went-web/security"
)

// HomeController example controller
type HomeController struct {
	controller.Base
}

func (c *HomeController) GetConfiguration() controller.Configuration {
	return controller.Configuration{
		Name: "Home",
		Routes: []controller.Route{
			{Path: "/", Method: "GET", MethodName: "Index", Handler: c.Index},
			{Path: "/about", Method: "GET", MethodName: "About", Handler: c.About},
			{Path: "/contact", Method: "GET", MethodName: "Contact", Handler: c.Contact},
		},
	}
}

func (c *HomeController) Index(ctx iris.Context) {
	ctx.WriteString("Welcome to Home Page!")
}

func (c *HomeController) About(ctx iris.Context) {
	ctx.WriteString("About Us Page")
}

func (c *HomeController) Contact(ctx iris.Context) {
	ctx.WriteString("Contact Page")
}

// APIController example controller
type APIController struct {
	controller.Base
}

func (c *APIController) GetConfiguration() controller.Configuration {
	return controller.Configuration{
		Name: "API",
		Routes: []controller.Route{
			{Path: "/api/data", Method: "GET", MethodName: "GetData", Handler: c.GetData},
			{Path: "/api/data", Method: "POST", MethodName: "PostData", Handler: c.PostData},
		},
	}
}

func (c *APIController) GetData(ctx iris.Context) {
	ctx.JSON(iris.Map{"data": []string{"item1", "item2", "item3"}})
}

func (c *APIController) PostData(ctx iris.Context) {
	ctx.JSON(iris.Map{"status": "created"})
}

// Example 7: Controllers with Registry
func main() {
	app := iris.New()

	// Create security (basic)
	sec, err := security.NewSecurity("config/security.yaml")
	if err != nil {
		// If no security.yaml, create minimal config
		sec = nil
	}

	// Create registry and register controllers
	registry := controller.NewRegistry()
	registry.Register(&HomeController{})
	registry.Register(&APIController{})

	// Register handlers with security if available
	if sec != nil {
		for _, h := range registry.Handlers() {
			sec.RegisterHandler(h.Controller, h.Method, h.Handler)
		}
	}

	// Register routes manually (without security)
	homeConfig := (&HomeController{}).GetConfiguration()
	for _, route := range homeConfig.Routes {
		app.Handle(route.Method, route.Path, route.Handler)
	}

	apiConfig := (&APIController{}).GetConfiguration()
	for _, route := range apiConfig.Routes {
		app.Handle(route.Method, route.Path, route.Handler)
	}

	log.Println("Server starting on :8086")
	log.Println("Test: curl <http://localhost:8086/>")
	log.Println("Test: curl <http://localhost:8086/about>")
	log.Println("Test: curl <http://localhost:8086/api/data>")
	app.Listen(":8086")
}
