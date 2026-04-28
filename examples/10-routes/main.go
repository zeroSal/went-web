package main

import (
	"log"

	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/security"
)

// Example 10: Route Configuration
func main() {
	app := iris.New()

	// Load security config
	sec, err := security.NewSecurity("config/security.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Load routes config
	routes, err := security.LoadRoutesConfig("config/routes.yaml")
	if err != nil {
		log.Fatal(err)
	}
	sec.SetRoutes(routes)

	// Register handlers
	sec.RegisterHandler("Home", "GET", func(ctx iris.Context) {
		ctx.WriteString("Home Page")
	})

	sec.RegisterHandler("Auth", "GET", func(ctx iris.Context) {
		ctx.WriteString("Login Page")
	})

	sec.RegisterHandler("Auth", "POST", func(ctx iris.Context) {
		ctx.WriteString("Login Handler")
	})

	sec.RegisterHandler("API", "GET", func(ctx iris.Context) {
		ctx.JSON(iris.Map{"data": []string{"item1", "item2"}})
	})

	sec.RegisterHandler("API", "POST", func(ctx iris.Context) {
		ctx.JSON(iris.Map{"status": "created"})
	})

	// Apply security middleware
	app.Use(sec.Middleware())

	// Register routes
	sec.RegisterRoutes(app)

	log.Println("Server starting on :8090")
	log.Println("Test: curl <http://localhost:8090/>")
	log.Println("Test: curl <http://localhost:8090/api/data> (requires auth)")
	app.Listen(":8090")
}
