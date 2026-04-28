package main

import (
	"log"

	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/security"
)

// Example 11: Session Management
func main() {
	app := iris.New()

	// Load security config with session
	sec, err := security.NewSecurity("config/security.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Get session manager
	session := sec.Session()

	// Login - set session through security
	app.Post("/login", func(ctx iris.Context) {
		// Get session from security
		s := session.Start(ctx)

		// In a real app, validate credentials here
		// For this example, we'll set user data
		s.Set("user_id", "123")
		s.Set("username", "john")

		// Set the session cookie that the authenticator expects
		ctx.SetCookie(&iris.Cookie{
			Name:     "SESSION_ID",
			Value:    s.ID(),
			HttpOnly: true,
		})

		ctx.JSON(iris.Map{"message": "Logged in!", "session": "active"})
	})

	// Check session
	app.Get("/profile", sec.Middleware(), func(ctx iris.Context) {
		s := session.Start(ctx)

		userID := s.GetString("user_id")
		username := s.GetString("username")

		ctx.JSON(iris.Map{
			"user_id":  userID,
			"username": username,
			"message":  "Session is active",
		})
	})

	// Logout - destroy session
	app.Get("/logout", func(ctx iris.Context) {
		s := session.Start(ctx)
		s.Clear()
		session.Destroy(ctx)

		ctx.JSON(iris.Map{"message": "Logged out!"})
	})

	log.Println("Server starting on :8091")
	log.Println("Login: curl -X POST <http://localhost:8091/login>")
	log.Println("Profile: curl -b cookies.txt -c cookies.txt <http://localhost:8091/profile>")
	log.Println("Logout: curl <http://localhost:8091/logout>")
	app.Listen(":8091")
}
