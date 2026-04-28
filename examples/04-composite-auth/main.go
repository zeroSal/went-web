package main

import (
	"log"

	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/auth"
)

// Example 4: Composite Authentication (try multiple methods)
func main() {
	app := iris.New()

	// Create multiple authenticators
	cookieAuth := auth.NewCookie("SESSION_ID")
	bearerAuth := auth.NewBearer()

	// Create composite - tries Cookie first, then Bearer
	compositeAuth := auth.NewComposite(
		cookieAuth,
		bearerAuth,
	)

	// Login with cookie
	app.Get("/login-cookie", func(ctx iris.Context) {
		ctx.SetCookie(&iris.Cookie{
			Name:  "SESSION_ID",
			Value: "cookie-session-123",
		})
		ctx.WriteString("Logged in with cookie!")
	})

	// Login with bearer token
	app.Get("/login-bearer", func(ctx iris.Context) {
		ctx.JSON(iris.Map{
			"token": "bearer-token-456",
			"message": "Use this token in Authorization header",
		})
	})

	// Protected route - works with either cookie OR bearer token
	app.Get("/api/data", compositeAuth.Middleware(), func(ctx iris.Context) {
		ctx.JSON(iris.Map{
			"message": "You are authenticated!",
			"method":  "Works with both cookie and bearer token",
		})
	})

	// Manual authentication check
	app.Get("/check", func(ctx iris.Context) {
		if user, ok := compositeAuth.Authenticate(ctx); ok {
			ctx.JSON(iris.Map{
				"authenticated": true,
				"user":           user,
			})
		} else {
			ctx.JSON(iris.Map{"authenticated": false})
		}
	})

	log.Println("Server starting on :8083")
	log.Println("Test cookie: curl <http://localhost:8083/login-cookie> then curl <http://localhost:8083/api/data> --cookie 'SESSION_ID=cookie-session-123'")
	log.Println("Test bearer: curl -H 'Authorization: Bearer bearer-token-456' <http://localhost:8083/api/data>")
	app.Listen(":8083")
}
