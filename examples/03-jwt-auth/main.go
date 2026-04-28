package main

import (
	"log"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/auth"
	"github.com/zeroSal/went-web/security"
)

// Example 3: JWT Authentication
func main() {
	app := iris.New()

	// Load security config from file
	sec, err := security.NewSecurity("config/security.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Get JWT authenticator from config
	jwtAuth := sec.Authenticator().(*auth.JWT)

	// Login route - generates JWT token
	app.Post("/login", func(ctx iris.Context) {
		username := ctx.PostValue("username")
		password := ctx.PostValue("password")

		// Simulate authentication
		if username == "admin" && password == "secret" {
			claims := jwt.MapClaims{
				"sub":      "user123",
				"username": username,
				"roles":    []string{"admin", "user"},
			}

			token, err := jwtAuth.GenerateToken(claims)
			if err != nil {
				ctx.StatusCode(500)
				ctx.JSON(iris.Map{"error": "Failed to generate token"})
				return
			}

			ctx.JSON(iris.Map{"token": token})
		} else {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{"error": "Invalid credentials"})
		}
	})

	// Protected route - validates JWT via middleware
	app.Get("/api/data", sec.Middleware(), func(ctx iris.Context) {
		if claims, ok := jwtAuth.Authenticate(ctx); ok {
			mapClaims := claims.(jwt.MapClaims)
			ctx.JSON(iris.Map{
				"message":  "Protected data",
				"user_id":  mapClaims["sub"],
				"username": mapClaims["username"],
			})
		} else {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{"error": "Unauthorized"})
		}
	})

	// Alternative: using middleware only
	app.Get("/api/profile", sec.Middleware(), func(ctx iris.Context) {
		ctx.WriteString("Profile page - you are authenticated!")
	})

	log.Println("Server starting on :8082")
	log.Println("1. Login: curl -X POST <http://localhost:8082/login> -d 'username=admin&password=secret'")
	log.Println("2. Use token: curl -H 'Authorization: Bearer <token>' <http://localhost:8082/api/data>")
	app.Listen(":8082")
}
