package main

import (
	"log"

	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/auth"
)

// Example 2: Bearer Authentication
func main() {
	app := iris.New()

	// Create bearer authenticator
	bearerAuth := auth.NewBearer()

	// Login route - returns token (simulated)
	app.Post("/login", func(ctx iris.Context) {
		username := ctx.PostValue("username")
		password := ctx.PostValue("password")

		// Simulate authentication
		if username == "admin" && password == "secret" {
			// In real app, generate a proper token
			token := "my-secret-token-123"
			ctx.JSON(iris.Map{"token": token})
		} else {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{"error": "Invalid credentials"})
		}
	})

	// Protected route - expects "Authorization: Bearer <token>"
	app.Get("/api/data", bearerAuth.Middleware(), func(ctx iris.Context) {
		if user, ok := bearerAuth.Authenticate(ctx); ok {
			ctx.JSON(iris.Map{
				"message": "Protected data",
				"token":   user,
			})
		} else {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{"error": "Unauthorized"})
		}
	})

	// Alternative: using middleware
	app.Get("/api/profile", bearerAuth.Middleware(), func(ctx iris.Context) {
		ctx.WriteString("Profile page - you are authenticated!")
	})

	log.Println("Server starting on :8081")
	log.Println("Test with: curl -H 'Authorization: Bearer my-secret-token-123' <http://localhost:8081/api/data>")
	app.Listen(":8081")
}
