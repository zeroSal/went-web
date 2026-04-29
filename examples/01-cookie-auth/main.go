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

// MockUserProvider maps session tokens to users
type MockUserProvider struct{}

func (p *MockUserProvider) Load(params any) (user.Interface, error) {
	token := params.(string)
	users := map[string]*User{
		"user-session-123": {
			ID:       "user123",
			Username: "testuser",
			Roles:    []string{"user"},
		},
		"admin-session-456": {
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

// Example 1: Cookie Authentication - Complete Test Cases
func main() {
	app := iris.New()

	// Create cookie authenticator
	cookieAuth := auth.NewCookie("SESSION_ID")
	cookieAuth.SetSessionProvider(&MockUserProvider{})

	// ========================================
	// AUTHENTICATION TEST ENDPOINTS
	// ========================================

	// Login - sets valid cookie (Authentication Success)
	app.Get("/login", func(ctx iris.Context) {
		ctx.SetCookie(&iris.Cookie{
			Name:     "SESSION_ID",
			Value:    "user-session-123",
			HttpOnly: true,
		})
		ctx.JSON(iris.Map{
			"message": "Logged in! Cookie set.",
			"cookie":  "SESSION_ID=user-session-123",
		})
	})

	// Login as admin - sets admin cookie (Authentication Success)
	app.Get("/login/admin", func(ctx iris.Context) {
		ctx.SetCookie(&iris.Cookie{
			Name:     "SESSION_ID",
			Value:    "admin-session-456",
			HttpOnly: true,
		})
		ctx.JSON(iris.Map{
			"message": "Logged in as admin! Cookie set.",
			"cookie":  "SESSION_ID=admin-session-456",
		})
	})

	// Test authentication without cookie (Authentication Failure - No Cookie)
	app.Get("/auth/no-cookie", func(ctx iris.Context) {
		if u, ok := cookieAuth.Authenticate(ctx); ok && u != nil {
			ctx.JSON(iris.Map{
				"authenticated": true,
				"username":      u.GetUsername(),
			})
		} else {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{
				"authenticated": false,
				"error":         "No cookie provided",
				"message":       "Authentication failed - no cookie",
			})
		}
	})

	// Test authentication with invalid cookie (Authentication Failure - Invalid Cookie)
	app.Get("/auth/invalid-cookie", func(ctx iris.Context) {
		ctx.SetCookie(&iris.Cookie{
			Name:  "SESSION_ID",
			Value: "invalid-session-id",
		})
		ctx.JSON(iris.Map{
			"message": "Cookie set to invalid value. Now test /profile to see failure.",
			"cookie":  "SESSION_ID=invalid-session-id",
		})
	})

	// ========================================
	// PROTECTED ROUTES - ACCESS GRANTED/DENIED
	// ========================================

	// Protected route - Access Granted (Valid Session)
	app.Get("/profile", cookieAuth.Middleware(), func(ctx iris.Context) {
		if u, ok := cookieAuth.Authenticate(ctx); ok && u != nil {
			ctx.JSON(iris.Map{
				"message":  "Welcome! Access granted.",
				"username": u.GetUsername(),
				"roles":    u.GetRoles(),
				"session":  "valid",
			})
		}
	})

	// Protected route - Access Denied (Invalid Session)
	// This will be handled by Middleware() automatically returning 401
	// Test with: curl --cookie 'SESSION_ID=invalid-session-id' <http://localhost:8080/profile>

	// Protected route - Access Denied (No Session)
	// Test with: curl <http://localhost:8080/profile> (no cookie)
	// Middleware will return 401 Unauthorized

	// Admin-only route - Access Granted (Valid Admin Session)
	app.Get("/admin", cookieAuth.Middleware(), func(ctx iris.Context) {
		if u, ok := cookieAuth.Authenticate(ctx); ok && u != nil {
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

	// Logout - Clear Session
	app.Get("/logout", func(ctx iris.Context) {
		ctx.RemoveCookie("SESSION_ID")
		ctx.JSON(iris.Map{
			"message": "Logged out! Cookie removed.",
		})
	})

	// ========================================
	// TEST INSTRUCTIONS
	// ========================================
	log.Println("==========================================")
	log.Println("Cookie Authentication Example - Running on :8080")
	log.Println("==========================================")
	log.Println("")
	log.Println("AUTHENTICATION SUCCESS:")
	log.Println("  curl <http://localhost:8080/login>")
	log.Println("  curl --cookie 'SESSION_ID=user-session-123' <http://localhost:8080/profile>")
	log.Println("")
	log.Println("AUTHENTICATION FAILURE - NO COOKIE:")
	log.Println("  curl <http://localhost:8080/auth/no-cookie>")
	log.Println("  curl <http://localhost:8080/profile>")
	log.Println("")
	log.Println("AUTHENTICATION FAILURE - INVALID COOKIE:")
	log.Println("  curl <http://localhost:8080/auth/invalid-cookie>")
	log.Println("  curl --cookie 'SESSION_ID=invalid-session-id' <http://localhost:8080/profile>")
	log.Println("")
	log.Println("PROTECTED ROUTE - ACCESS GRANTED:")
	log.Println("  curl --cookie 'SESSION_ID=user-session-123' <http://localhost:8080/profile>")
	log.Println("")
	log.Println("PROTECTED ROUTE - ACCESS DENIED:")
	log.Println("  curl <http://localhost:8080/profile> (no cookie - 401)")
	log.Println("  curl --cookie 'SESSION_ID=invalid' <http://localhost:8080/profile> (invalid - 401)")
	log.Println("")
	log.Println("ADMIN ROUTE - ACCESS GRANTED:")
	log.Println("  curl <http://localhost:8080/login/admin>")
	log.Println("  curl --cookie 'SESSION_ID=admin-session-456' <http://localhost:8080/admin>")
	log.Println("")
	log.Println("ADMIN ROUTE - ACCESS DENIED (Forbidden):")
	log.Println("  curl --cookie 'SESSION_ID=user-session-123' <http://localhost:8080/admin>")
	log.Println("")
	log.Println("LOGOUT:")
	log.Println("  curl --cookie 'SESSION_ID=user-session-123' <http://localhost:8080/logout>")
	log.Println("==========================================")

	app.Listen(":8080")
}

// Ensure user.Interface is used
var _ user.Interface = (*User)(nil)
