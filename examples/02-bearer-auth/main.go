package main

import (
	"log"
	"slices"

	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/auth"
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

// MockUserProvider maps bearer tokens to users
type MockUserProvider struct{}

func (p *MockUserProvider) Load(params any) (user.Interface, error) {
	token := params.(string)
	users := map[string]*User{
		"my-secret-token-123": {
			ID:       "user123",
			Username: "testuser",
			Roles:    []string{"user"},
		},
		"admin-token-456": {
			ID:       "admin456",
			Username: "admin",
			Roles:    []string{"admin", "user"},
		},
	}
	if u, ok := users[token]; ok {
		return u, nil
	}
	return nil, nil
}

// Example 2: Bearer Authentication - Complete Test Cases
func main() {
	app := iris.New()

	// Create bearer authenticator
	bearerAuth := auth.NewBearer()
	bearerAuth.SetSessionProvider(&MockUserProvider{})

	// ========================================
	// AUTHENTICATION TEST ENDPOINTS
	// ========================================

	// Login - returns valid token (Authentication Success)
	app.Get("/login", func(ctx iris.Context) {
		ctx.JSON(iris.Map{
			"token":    "my-secret-token-123",
			"username": "testuser",
			"roles":    []string{"user"},
			"message":  "Login successful! Use this token in Authorization header.",
		})
	})

	app.Post("/login", func(ctx iris.Context) {
		username := ctx.PostValue("username")
		password := ctx.PostValue("password")

		// Simulate authentication
		if username == "admin" && password == "secret" {
			token := "my-secret-token-123"
			ctx.JSON(iris.Map{
				"token":   token,
				"message": "Login successful! Use this token in Authorization header.",
			})
		} else if username == "testuser" && password == "secret" {
			token := "my-secret-token-123"
			ctx.JSON(iris.Map{
				"token":   token,
				"message": "Login successful! Use this token in Authorization header.",
			})
		} else {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{
				"error":   "Invalid credentials",
				"message": "Authentication failed",
			})
		}
	})

	// Test authentication without token (Authentication Failure - No Token)
	app.Get("/auth/no-token", func(ctx iris.Context) {
		if u, ok := bearerAuth.Authenticate(ctx); ok && u != nil {
			ctx.JSON(iris.Map{
				"authenticated": true,
				"username":      u.GetUsername(),
			})
		} else {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{
				"authenticated": false,
				"error":         "No token provided",
				"message":       "Authentication failed - no Bearer token",
			})
		}
	})

	// Test authentication with invalid token (Authentication Failure - Invalid Token)
	app.Get("/auth/invalid-token", func(ctx iris.Context) {
		if u, ok := bearerAuth.Authenticate(ctx); ok && u != nil {
			ctx.JSON(iris.Map{
				"authenticated": true,
				"username":      u.GetUsername(),
			})
		} else {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{
				"authenticated": false,
				"error":         "Invalid token",
				"message":       "Authentication failed - invalid Bearer token",
			})
		}
	})

	// ========================================
	// PROTECTED ROUTES - ACCESS GRANTED/DENIED
	// ========================================

	// Protected route - Access Granted (Valid Token)
	app.Get("/api/data", bearerAuth.Middleware(), func(ctx iris.Context) {
		if u, ok := bearerAuth.Authenticate(ctx); ok && u != nil {
			ctx.JSON(iris.Map{
				"message":      "Protected data - Access granted.",
				"username":     u.GetUsername(),
				"roles":        u.GetRoles(),
				"token_status": "valid",
			})
		}
	})

	// Protected route - Access Denied (Invalid/No Token)
	// Middleware automatically returns 401 for invalid or missing tokens
	// Test with: curl <http://localhost:8081/api/data> (no token)
	// Test with: curl -H "Authorization: Bearer invalid-token" <http://localhost:8081/api/data>

	// Admin-only route - Access Granted (Valid Admin Token)
	app.Get("/api/admin", bearerAuth.Middleware(), func(ctx iris.Context) {
		if u, ok := bearerAuth.Authenticate(ctx); ok && u != nil {
			if u.HasRole("admin") {
				ctx.JSON(iris.Map{
					"message":  "Welcome Admin! Access granted.",
					"username": u.GetUsername(),
					"roles":    u.GetRoles(),
				})
			} else {
				ctx.StatusCode(403)
				ctx.JSON(iris.Map{
					"error":   "Forbidden",
					"message": "Admin access required",
				})
			}
		}
	})

	// ========================================
	// TEST INSTRUCTIONS
	// ========================================
	log.Println("==========================================")
	log.Println("Bearer Authentication Example - Running on :8081")
	log.Println("==========================================")
	log.Println("")
	log.Println("AUTHENTICATION SUCCESS:")
	log.Println("  curl -X POST -d 'username=testuser&password=secret' <http://localhost:8081/login>")
	log.Println("  curl -H 'Authorization: Bearer my-secret-token-123' <http://localhost:8081/api/data>")
	log.Println("")
	log.Println("AUTHENTICATION FAILURE - NO TOKEN:")
	log.Println("  curl <http://localhost:8081/auth/no-token>")
	log.Println("  curl <http://localhost:8081/api/data>")
	log.Println("")
	log.Println("AUTHENTICATION FAILURE - INVALID TOKEN:")
	log.Println("  curl -H 'Authorization: Bearer invalid-token' <http://localhost:8081/auth/invalid-token>")
	log.Println("  curl -H 'Authorization: Bearer invalid-token' <http://localhost:8081/api/data>")
	log.Println("")
	log.Println("PROTECTED ROUTE - ACCESS GRANTED:")
	log.Println("  curl -H 'Authorization: Bearer my-secret-token-123' <http://localhost:8081/api/data>")
	log.Println("")
	log.Println("PROTECTED ROUTE - ACCESS DENIED:")
	log.Println("  curl <http://localhost:8081/api/data> (no token - 401)")
	log.Println("  curl -H 'Authorization: Bearer invalid' <http://localhost:8081/api/data> (invalid - 401)")
	log.Println("")
	log.Println("ADMIN ROUTE - ACCESS GRANTED:")
	log.Println("  curl -X POST -d 'username=admin&password=admin123' <http://localhost:8081/login>")
	log.Println("  curl -H 'Authorization: Bearer admin-token-456' <http://localhost:8081/api/admin>")
	log.Println("")
	log.Println("ADMIN ROUTE - ACCESS DENIED (Forbidden):")
	log.Println("  curl -H 'Authorization: Bearer my-secret-token-123' <http://localhost:8081/api/admin>")
	log.Println("==========================================")

	app.Listen(":8081")
}

// Ensure user.Interface is used
var _ user.Interface = (*User)(nil)
