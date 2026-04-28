package main

import (
	"log"

	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/security"
)

// Example 12: CSRF Protection
func main() {
	app := iris.New()

	// Load security config with CSRF
	sec, err := security.NewSecurity("config/security.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Get CSRF middleware
	csrfMiddleware := sec.CSRF()

	// Public route - get CSRF token
	app.Get("/csrf-token", func(ctx iris.Context) {
		ctx.JSON(iris.Map{
			"csrf_token": "CSRF token would be generated here",
			"note":       "In production, use csrf.CreateToken(ctx)",
		})
	})

	// Form page (no CSRF check for GET)
	app.Get("/form", func(ctx iris.Context) {
		ctx.HTML(`
			<form method="POST" action="/form">
				<input type="hidden" name="csrf_token" value="TOKEN_HERE">
				<input name="data" placeholder="Enter data">
				<button>Submit</button>
			</form>
		`)
	})

	// Protected POST - requires CSRF token
	app.Post("/form", csrfMiddleware, func(ctx iris.Context) {
		data := ctx.PostValue("data")
		ctx.JSON(iris.Map{
			"message": "Form submitted successfully!",
			"data":    data,
		})
	})

	// Another protected route with CSRF
	app.Delete("/api/resource", csrfMiddleware, func(ctx iris.Context) {
		ctx.JSON(iris.Map{"message": "Resource deleted"})
	})

	// Check if CSRF is enabled
	app.Get("/check-csrf", func(ctx iris.Context) {
		if sec.CSRF() != nil {
			ctx.JSON(iris.Map{
				"csrf_enabled": true,
				"header_name":  "X-CSRF-Token",
				"field_name":   "csrf_token",
			})
		} else {
			ctx.JSON(iris.Map{"csrf_enabled": false})
		}
	})

	log.Println("Server starting on :8092")
	log.Println("Get CSRF token: curl <http://localhost:8092/csrf-token>")
	log.Println("Form page: <http://localhost:8092/form>")
	log.Println("POST requires CSRF token in header or form field")
	app.Listen(":8092")
}
