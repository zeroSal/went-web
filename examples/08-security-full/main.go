package main

import (
	"fmt"
	"log"
	"slices"

	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/security"
	"github.com/zeroSal/went-web/session"
	"github.com/zeroSal/went-web/user"
)

// User implements user.Interface
type User struct {
	ID       any
	Username string
	Roles    []string
}

func (u *User) GetID() any {
	return u.ID
}

func (u *User) GetUsername() string {
	return u.Username
}

func (u *User) GetRoles() []string {
	return u.Roles
}

func (u *User) HasRole(role string) bool {
	return slices.Contains(u.Roles, role)
}

// MockUserProvider maps session tokens to users
type MockUserProvider struct{}

func (p *MockUserProvider) Load(credential any) (user.Interface, error) {
	token := credential.(string)
	users := map[string]*User{
		"session-admin-123": {
			ID:       "1",
			Username: "admin",
			Roles:    []string{"admin"},
		},
		"session-user-456": {
			ID:       "2",
			Username: "john",
			Roles:    []string{"user"},
		},
	}
	if u, ok := users[token]; ok {
		return u, nil
	}
	return nil, fmt.Errorf("user not found")
}

// Example 8: Full Security with YAML config - Complete Test Cases
func main() {
	app := iris.New()

	// Load security config from file
	sec, err := security.NewSecurity("config/security.yaml")
	if err != nil {
		log.Fatalf("Failed to load security config: %v", err)
	}

	// Set session provider
	sec.SetSessionProvider(&MockUserProvider{})

	// ========================================
	// AUTHENTICATION SUCCESS
	// ========================================

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

	// Login handler - sets session cookie (Authentication Success)
	app.Post("/login", func(ctx iris.Context) {
		username := ctx.PostValue("username")
		password := ctx.PostValue("password")

		// Simple auth check
		if username == "admin" && password == "secret" {
			// Set session cookie with valid token
			ctx.SetCookie(&iris.Cookie{
				Name:     "SESSION_ID",
				Value:    "session-admin-123",
				HttpOnly: true,
			})

			ctx.JSON(iris.Map{
				"message": "Logged in as admin!",
				"cookie":  "SESSION_ID=session-admin-123",
			})
		} else if username == "john" && password == "doe" {
			ctx.SetCookie(&iris.Cookie{
				Name:     "SESSION_ID",
				Value:    "session-user-456",
				HttpOnly: true,
			})

			ctx.JSON(iris.Map{
				"message": "Logged in as user!",
				"cookie":  "SESSION_ID=session-user-456",
			})
		} else {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{"error": "Invalid credentials"})
		}
	})

	// Login via GET for easier testing (Authentication Success)
	app.Get("/login/{username}", func(ctx iris.Context) {
		username := ctx.Params().Get("username")

		if username == "admin" {
			ctx.SetCookie(&iris.Cookie{
				Name:     "SESSION_ID",
				Value:    "session-admin-123",
				HttpOnly: true,
			})
			ctx.JSON(iris.Map{
				"message": "Logged in as admin (GET)!",
				"cookie":  "SESSION_ID=session-admin-123",
			})
		} else if username == "john" {
			ctx.SetCookie(&iris.Cookie{
				Name:     "SESSION_ID",
				Value:    "session-user-456",
				HttpOnly: true,
			})
			ctx.JSON(iris.Map{
				"message": "Logged in as john (GET)!",
				"cookie":  "SESSION_ID=session-user-456",
			})
		} else {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{"error": "Invalid username"})
		}
	})

	// ========================================
	// AUTHENTICATION FAILURE
	// ========================================

	// Test with invalid credentials (Authentication Failure)
	app.Get("/login/invalid", func(ctx iris.Context) {
		ctx.StatusCode(401)
		ctx.JSON(iris.Map{
			"error":   "Invalid credentials",
			"message": "Authentication failed",
		})
	})

	// Test with invalid session cookie (Authentication Failure)
	app.Get("/auth/invalid-session", func(ctx iris.Context) {
		ctx.SetCookie(&iris.Cookie{
			Name:  "SESSION_ID",
			Value: "invalid-session-id",
		})
		ctx.JSON(iris.Map{
			"message": "Cookie set to invalid value. Now test /profile to see 401.",
			"cookie":  "SESSION_ID=invalid-session-id",
		})
	})

	// ========================================
	// PROTECTED ROUTES - ACCESS GRANTED
	// ========================================

	// Protected route using security middleware - Access Granted
	app.Get("/admin", sec.Middleware(), func(ctx iris.Context) {
		// Get authenticated user
		if auth := sec.Authenticator(); auth != nil {
			if u, ok := auth.Authenticate(ctx); ok && u != nil {
				ctx.JSON(iris.Map{
					"message":  "Welcome to Admin Panel! Access granted.",
					"username": u.GetUsername(),
					"roles":    u.GetRoles(),
				})
				return
			}
		}
		ctx.JSON(iris.Map{"message": "Welcome to Admin Panel!"})
	})

	// Another protected route - Access Granted
	app.Get("/profile", sec.Middleware(), func(ctx iris.Context) {
		if auth := sec.Authenticator(); auth != nil {
			if u, ok := auth.Authenticate(ctx); ok && u != nil {
				ctx.JSON(iris.Map{
					"message":  "Your Profile Page. Access granted.",
					"username": u.GetUsername(),
					"roles":    u.GetRoles(),
				})
				return
			}
		}
		ctx.WriteString("Your Profile Page")
	})

	// ========================================
	// PROTECTED ROUTES - ACCESS DENIED
	// ========================================
	// Test with no auth: curl <http://localhost:8088/admin> (401)
	// Test with invalid auth: curl --cookie "SESSION_ID=invalid" <http://localhost:8088/admin> (401)

	// Logout
	app.Get("/logout", func(ctx iris.Context) {
		ctx.RemoveCookie("SESSION_ID")
		ctx.JSON(iris.Map{"message": "Logged out!"})
	})

	// ========================================
	// TEST INSTRUCTIONS
	// ========================================
	log.Println("==========================================")
	log.Println("Full Security Example - Running on :8088")
	log.Println("==========================================")
	log.Println("")
	log.Println("AUTHENTICATION SUCCESS:")
	log.Println("  curl <http://localhost:8088/login/admin>")
	log.Println("  curl --cookie 'SESSION_ID=session-admin-123' <http://localhost:8088/admin>")
	log.Println("")
	log.Println("AUTHENTICATION FAILURE - NO SESSION:")
	log.Println("  curl <http://localhost:8088/admin> (401)")
	log.Println("  curl <http://localhost:8088/profile> (401)")
	log.Println("")
	log.Println("AUTHENTICATION FAILURE - INVALID SESSION:")
	log.Println("  curl <http://localhost:8088/auth/invalid-session>")
	log.Println("  curl --cookie 'SESSION_ID=invalid-session-id' <http://localhost:8088/admin> (401)")
	log.Println("")
	log.Println("PROTECTED ROUTE - ACCESS GRANTED:")
	log.Println("  curl --cookie 'SESSION_ID=session-admin-123' <http://localhost:8088/admin>")
	log.Println("  curl --cookie 'SESSION_ID=session-user-456' <http://localhost:8088/profile>")
	log.Println("")
	log.Println("PROTECTED ROUTE - ACCESS DENIED:")
	log.Println("  curl <http://localhost:8088/admin> (no session - 401)")
	log.Println("  curl --cookie 'SESSION_ID=invalid' <http://localhost:8088/admin> (invalid - 401)")
	log.Println("")
	log.Println("LOGOUT:")
	log.Println("  curl --cookie 'SESSION_ID=session-admin-123' <http://localhost:8088/logout>")
	log.Println("==========================================")

	app.Listen(":8088")
}

// Ensure interfaces are used
var _ session.ProviderInterface = (*MockUserProvider)(nil)
var _ user.Interface = (*User)(nil)
