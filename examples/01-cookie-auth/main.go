package main

import (
	"log"

	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/auth"
)

// Example 1: Cookie Authentication
func main() {
	app := iris.New()

	// Create cookie authenticator
	cookieAuth := auth.NewCookie("SESSION_ID")

	// Public route
	app.Get("/login", func(ctx iris.Context) {
		// Simulate login - set cookie
		ctx.SetCookie(&iris.Cookie{
			Name:  "SESSION_ID",
			Value: "user-session-123",
		})
		ctx.WriteString("Logged in! Cookie set.")
	})

	// Protected route using cookie auth
	app.Get("/profile", cookieAuth.Middleware(), func(ctx iris.Context) {
		if user, ok := cookieAuth.Authenticate(ctx); ok {
			ctx.WriteString("Welcome! Your session: " + user.(string))
		} else {
			ctx.StatusCode(401)
			ctx.WriteString("Unauthorized")
		}
	})

	// Logout
	app.Get("/logout", func(ctx iris.Context) {
		ctx.RemoveCookie("SESSION_ID")
		ctx.WriteString("Logged out!")
	})

	log.Println("Server starting on :8080")
	log.Println("Test: curl <http://localhost:8080/login>")
	log.Println("Test: curl --cookie 'SESSION_ID=user-session-123' <http://localhost:8080/profile>")
	app.Listen(":8080")
}
