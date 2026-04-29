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
	}
}

func (c *APIController) GetData(ctx iris.Context) {
	ctx.JSON(iris.Map{"data": []string{"item1", "item2", "item3"}})
}

func (c *APIController) PostData(ctx iris.Context) {
	ctx.JSON(iris.Map{"status": "created"})
}

// Example 7: Controllers with Routes YAML
func main() {
	app := iris.New()

	// Load security config (for routes)
	sec, err := security.NewSecurity("config/security.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Create registry and register controllers
	registry := controller.NewRegistry()
	registry.Register(&HomeController{})
	registry.Register(&APIController{})

	// Register handlers with security
	for _, h := range registry.Handlers() {
		sec.RegisterHandler(h.Controller, h.Method, h.Handler)
	}

	// Load routes config
	routes, err := security.LoadRoutesConfig("config/routes.yaml")
	if err != nil {
		log.Fatal(err)
	}
	sec.SetRoutes(routes)

	// Register routes from YAML
	sec.RegisterRoutes(app)

	log.Println("Server starting on :8086")
	log.Println("Test: curl <http://localhost:8086/>")
	log.Println("Test: curl <http://localhost:8086/about>")
	log.Println("Test: curl <http://localhost:8086/api/data>")
	app.Listen(":8086")
}
