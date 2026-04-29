package main

import (
	"log"
	"slices"
	"time"

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

// MockUserProvider for this example
type MockUserProvider struct{}

func (p *MockUserProvider) Load(credential any) (user.Interface, error) {
	// Check by username
	if username, ok := credential.(string); ok {
		users := map[string]*User{
			"john": {
				ID:       "1",
				Username: "john",
				Roles:    []string{"user"},
			},
			"admin": {
				ID:       "2",
				Username: "admin",
				Roles:    []string{"admin", "user"},
			},
		}
		if u, ok := users[username]; ok {
			return u, nil
		}

		// Check by session token
		tokens := map[string]*User{
			"token-john-123": {
				ID:       "1",
				Username: "john",
				Roles:    []string{"user"},
			},
			"token-admin-456": {
				ID:       "2",
				Username: "admin",
				Roles:    []string{"admin", "user"},
			},
		}
		if u, ok := tokens[username]; ok {
			return u, nil
		}
	}

	return nil, nil
}

// Example 9: Complete Integration - Complete Test Cases
func main() {
	app := iris.New()

	// Load security config from file
	sec, err := security.NewSecurity("config/security.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Set session provider
	sec.SetSessionProvider(&MockUserProvider{})

	// ========================================
	// AUTHENTICATION SUCCESS
	// ========================================

	// Login endpoint - returns token (Authentication Success)
	app.Post("/login", func(ctx iris.Context) {
		username := ctx.PostValue("username")
		password := ctx.PostValue("password")

		// Simple check
		if (username == "john" || username == "admin") && password == "secret" {
			// Get user
			u, _ := (&MockUserProvider{}).Load(username)

			// Return appropriate token based on user
			var token string
			if username == "john" {
				token = "token-john-123"
			} else {
				token = "token-admin-456"
			}

			ctx.JSON(iris.Map{
				"token":    token,
				"username": u.GetUsername(),
				"roles":    u.GetRoles(),
				"message":  "Login successful",
			})
		} else {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{"error": "Invalid credentials"})
		}
	})

	// Login via GET for easier testing (Authentication Success)
	app.Get("/login/{username}", func(ctx iris.Context) {
		username := ctx.Params().Get("username")

		u, err := (&MockUserProvider{}).Load(username)
		if err != nil {
			ctx.StatusCode(404)
			ctx.JSON(iris.Map{"error": "User not found"})
			return
		}

		ctx.JSON(iris.Map{
			"username": u.GetUsername(),
			"roles":    u.GetRoles(),
			"message":  "Login successful (GET)",
		})
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

	// ========================================
	// PUBLIC ENDPOINTS
	// ========================================

	// Public endpoint
	app.Get("/public/info", func(ctx iris.Context) {
		ctx.JSON(iris.Map{
			"message": "This is public information",
			"time":    time.Now(),
		})
	})

	// ========================================
	// PROTECTED ENDPOINTS - ACCESS GRANTED
	// ========================================

	// Protected endpoint - Access Granted (with valid Bearer token)
	app.Get("/api/profile", sec.Middleware(), func(ctx iris.Context) {
		// Get user from authenticator
		if auth := sec.Authenticator(); auth != nil {
			if u, ok := auth.Authenticate(ctx); ok && u != nil {
				ctx.JSON(iris.Map{
					"user_id":  u.GetID(),
					"username": u.GetUsername(),
					"roles":    u.GetRoles(),
					"message":  "Access granted to profile",
				})
				return
			}
		}
		ctx.JSON(iris.Map{"message": "Your Profile Page"})
	})

	// Admin only endpoint - Access Granted (with admin token)
	app.Get("/admin/users", sec.Middleware(), func(ctx iris.Context) {
		// Check if user has admin role
		if auth := sec.Authenticator(); auth != nil {
			if u, ok := auth.Authenticate(ctx); ok && u != nil {
				if u.HasRole("admin") || u.HasRole("ROLE_admin") {
					ctx.JSON(iris.Map{
						"users":   []string{"john", "admin"},
						"count":   2,
						"message": "Access granted to admin endpoint",
					})
				} else {
					ctx.StatusCode(403)
					ctx.JSON(iris.Map{"error": "Forbidden: Admin access required"})
				}
				return
			}
		}
		ctx.StatusCode(401)
		ctx.JSON(iris.Map{"error": "Unauthorized"})
	})

	// ========================================
	// PROTECTED ENDPOINTS - ACCESS DENIED
	// ========================================
	// Test with no auth: curl <http://localhost:8089/api/profile> (401)
	// Test with invalid auth: curl -H "Authorization: Bearer invalid" <http://localhost:8089/api/profile> (401)
	// Test non-admin: curl -H "Authorization: Bearer token-john-123" <http://localhost:8089/admin/users> (403)

	// ========================================
	// TEST INSTRUCTIONS
	// ========================================
	log.Println("==========================================")
	log.Println("Integration Example - Running on :8089")
	log.Println("==========================================")
	log.Println("")
	log.Println("AUTHENTICATION SUCCESS:")
	log.Println("  curl -X POST -d 'username=john&password=secret' <http://localhost:8089/login>")
	log.Println("  curl -X POST -d 'username=admin&password=secret' <http://localhost:8089/login>")
	log.Println("")
	log.Println("AUTHENTICATION FAILURE:")
	log.Println("  curl -X POST -d 'username=invalid&password=wrong' <http://localhost:8089/login>")
	log.Println("  curl <http://localhost:8089/login/invalid>")
	log.Println("")
	log.Println("PUBLIC ENDPOINT:")
	log.Println("  curl <http://localhost:8089/public/info>")
	log.Println("")
	log.Println("PROTECTED ROUTE - ACCESS GRANTED:")
	log.Println("  curl -H 'Authorization: Bearer token-john-123' <http://localhost:8089/api/profile>")
	log.Println("")
	log.Println("PROTECTED ROUTE - ACCESS DENIED:")
	log.Println("  curl <http://localhost:8089/api/profile> (no token - 401)")
	log.Println("  curl -H 'Authorization: Bearer invalid' <http://localhost:8089/api/profile> (invalid - 401)")
	log.Println("")
	log.Println("ADMIN ROUTE - ACCESS GRANTED:")
	log.Println("  curl -H 'Authorization: Bearer token-admin-456' <http://localhost:8089/admin/users>")
	log.Println("")
	log.Println("ADMIN ROUTE - ACCESS DENIED (Forbidden):")
	log.Println("  curl -H 'Authorization: Bearer token-john-123' <http://localhost:8089/admin/users>")
	log.Println("==========================================")

	app.Listen(":8089")
}

// Ensure interfaces are used
var _ session.ProviderInterface = (*MockUserProvider)(nil)
var _ user.Interface = (*User)(nil)
