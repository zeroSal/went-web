package main

import (
	"log"

	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/auth"
	"github.com/zeroSal/went-web/security"
)

// Example 8: Full Security with YAML config
func main() {
	app := iris.New()

	// Load security config from file
	sec, err := security.NewSecurity("config/security.yaml")
	if err != nil {
		log.Fatalf("Failed to load security config: %v", err)
	}

	// Public route (no auth required)
	app.Get("/login", func(ctx iris.Context) {
		ctx.WriteString(`
			<form method="POST" action="/login">
				<input name="username" placeholder="Username">
				<input name="password" type="password" placeholder="Password">
				<button>Login</button>
			</form>
		`)
	})

	// Login handler
	app.Post("/login", func(ctx iris.Context) {
		username := ctx.PostValue("username")
		password := ctx.PostValue("password")

		// Simple auth check
		if username == "admin" && password == "secret" {
			// Generate JWT token
			jwtAuth := sec.Authenticator().(*auth.JWT)
			claims := map[string]interface{}{
				"sub":      "1",
				"username": username,
				"roles":    []string{"admin"},
			}
			token, _ := jwtAuth.GenerateToken(claims)

			// Set as cookie
			ctx.SetCookie(&iris.Cookie{
				Name:  "jwt",
				Value: token,
			})

			ctx.WriteString("Logged in! JWT set in cookie.")
		} else {
			ctx.StatusCode(401)
			ctx.WriteString("Invalid credentials")
		}
	})

	// Protected route using security middleware
	app.Get("/admin", sec.Middleware(), func(ctx iris.Context) {
		ctx.WriteString("Welcome to Admin Panel!")
	})

	// Another protected route
	app.Get("/profile", sec.Middleware(), func(ctx iris.Context) {
		ctx.WriteString("Your Profile Page")
	})

	// Logout
	app.Get("/logout", func(ctx iris.Context) {
		ctx.RemoveCookie("jwt")
		ctx.WriteString("Logged out!")
	})

	log.Println("Server starting on :8088")
	log.Println("1. Open <http://localhost:8088/login>")
	log.Println("2. Login with admin/secret")
	log.Println("3. Access <http://localhost:8088/admin> (protected)")
	app.Listen(":8088")
}
