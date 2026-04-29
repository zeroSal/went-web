package main

import (
	"log"

	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/security"
)

// Example 11: Session Management - Complete Test Cases
func main() {
	app := iris.New()

	// Load security config with session
	sec, err := security.NewSecurity("config/security.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Get session manager
	session := sec.GetSessionManager()

	// ========================================
	// AUTHENTICATION SUCCESS
	// ========================================

	// Login - set session through security (Authentication Success)
	app.Get("/login", func(ctx iris.Context) {
		s := session.Start(ctx)
		s.Set("user_id", "123")
		s.Set("username", "john")
		s.Set("role", "user")

		ctx.SetCookie(&iris.Cookie{
			Name:     "SESSION_ID",
			Value:    s.ID(),
			HttpOnly: true,
		})

		ctx.JSON(iris.Map{
			"message":  "Logged in!",
			"session":  "active",
			"username": "john",
		})
	})

	app.Post("/login", func(ctx iris.Context) {
		s := session.Start(ctx)
		s.Set("user_id", "123")
		s.Set("username", "john")
		s.Set("role", "user")

		ctx.SetCookie(&iris.Cookie{
			Name:     "SESSION_ID",
			Value:    s.ID(),
			HttpOnly: true,
		})

		ctx.JSON(iris.Map{
			"message":  "Logged in!",
			"session":  "active",
			"username": "john",
		})
	})

	// ========================================
	// AUTHENTICATION FAILURE
	// ========================================

	// Test without session (Authentication Failure - No Session)
	app.Get("/auth/none", func(ctx iris.Context) {
		s := session.Start(ctx)
		if s.GetString("user_id") == "" {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{
				"authenticated": false,
				"error":         "No session found",
				"message":       "Authentication failed - no valid session",
			})
			return
		}
		ctx.JSON(iris.Map{"authenticated": true})
	})

	// ========================================
	// PROTECTED ROUTES - ACCESS GRANTED
	// ========================================

	// Check session - Access Granted (Valid Session)
	app.Get("/profile", sec.Middleware(), func(ctx iris.Context) {
		s := session.Start(ctx)

		userID := s.GetString("user_id")
		username := s.GetString("username")
		role := s.GetString("role")

		ctx.JSON(iris.Map{
			"user_id":  userID,
			"username": username,
			"role":     role,
			"message":  "Session is active - Access granted",
		})
	})

	// Admin-only route - Access Granted (Valid Admin Session)
	app.Get("/admin", sec.Middleware(), func(ctx iris.Context) {
		s := session.Start(ctx)

		role := s.GetString("role")
		if role == "admin" {
			ctx.JSON(iris.Map{
				"message":  "Welcome Admin! Access granted.",
				"username": s.GetString("username"),
				"role":     role,
			})
		} else {
			ctx.StatusCode(403)
			ctx.JSON(iris.Map{
				"error":   "Forbidden",
				"message": "Admin access required",
			})
		}
	})

	// ========================================
	// PROTECTED ROUTES - ACCESS DENIED
	// ========================================
	// Test with no session: curl <http://localhost:8091/profile> (401)
	// Test with invalid session: curl -b "SESSION_ID=invalid" <http://localhost:8091/profile> (401)

	// ========================================
	// LOGOUT
	// ========================================

	// Logout - destroy session
	app.Get("/logout", func(ctx iris.Context) {
		s := session.Start(ctx)
		s.Clear()
		session.Destroy(ctx)

		ctx.JSON(iris.Map{"message": "Logged out!"})
	})

	// ========================================
	// TEST INSTRUCTIONS
	// ========================================
	log.Println("==========================================")
	log.Println("Session Management Example - Running on :8091")
	log.Println("==========================================")
	log.Println("")
	log.Println("AUTHENTICATION SUCCESS:")
	log.Println("  curl <http://localhost:8091/login>")
	log.Println("  curl --cookie 'SESSION_ID=<session-id>' <http://localhost:8091/profile>")
	log.Println("")
	log.Println("AUTHENTICATION SUCCESS - ADMIN:")
	log.Println("  curl <http://localhost:8091/login/admin>")
	log.Println("  curl --cookie 'SESSION_ID=<session-id>' <http://localhost:8091/admin>")
	log.Println("")
	log.Println("AUTHENTICATION FAILURE - NO SESSION:")
	log.Println("  curl <http://localhost:8091/auth/none>")
	log.Println("  curl <http://localhost:8091/profile> (401)")
	log.Println("")
	log.Println("AUTHENTICATION FAILURE - INVALID SESSION:")
	log.Println("  curl --cookie 'SESSION_ID=invalid-id' <http://localhost:8091/profile> (401)")
	log.Println("")
	log.Println("PROTECTED ROUTE - ACCESS GRANTED:")
	log.Println("  curl --cookie 'SESSION_ID=<valid-id>' <http://localhost:8091/profile>")
	log.Println("")
	log.Println("PROTECTED ROUTE - ACCESS DENIED:")
	log.Println("  curl <http://localhost:8091/profile> (no session - 401)")
	log.Println("")
	log.Println("ADMIN ROUTE - ACCESS GRANTED:")
	log.Println("  curl --cookie 'SESSION_ID=<admin-session-id>' <http://localhost:8091/admin>")
	log.Println("")
	log.Println("ADMIN ROUTE - ACCESS DENIED (Forbidden):")
	log.Println("  curl --cookie 'SESSION_ID=<user-session-id>' <http://localhost:8091/admin>")
	log.Println("")
	log.Println("LOGOUT:")
	log.Println("  curl --cookie 'SESSION_ID=<session-id>' <http://localhost:8091/logout>")
	log.Println("==========================================")

	app.Listen(":8091")
}
