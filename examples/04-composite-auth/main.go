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

// MockUserProvider maps both cookie and bearer tokens to users
type MockUserProvider struct{}

func (p *MockUserProvider) Load(params any) (user.Interface, error) {
	token := params.(string)
	users := map[string]*User{
		"cookie-session-123": {
			ID:       "user123",
			Username: "cookieuser",
			Roles:    []string{"user"},
		},
		"bearer-token-456": {
			ID:       "user456",
			Username: "beareruser",
			Roles:    []string{"user"},
		},
		"admin-cookie-789": {
			ID:       "admin123",
			Username: "admin",
			Roles:    []string{"admin", "user"},
		},
		"admin-bearer-000": {
			ID:       "admin456",
			Username: "adminbearer",
			Roles:    []string{"admin", "user"},
		},
	}
	if u, ok := users[token]; ok {
		return u, nil
	}
	return nil, nil
}

// Example 4: Composite Authentication - Complete Test Cases
// Tries Cookie first, then Bearer token
func main() {
	app := iris.New()

	// Create multiple authenticators
	cookieAuth := auth.NewCookie("SESSION_ID")
	bearerAuth := auth.NewBearer()

	// Set shared SessionProvider
	provider := &MockUserProvider{}
	cookieAuth.SetSessionProvider(provider)
	bearerAuth.SetSessionProvider(provider)

	// Create composite - tries Cookie first, then Bearer
	compositeAuth := auth.NewComposite(
		cookieAuth,
		bearerAuth,
	)

	// ========================================
	// AUTHENTICATION SUCCESS - COOKIE
	// ========================================

	// Login with cookie (Authentication Success - Cookie)
	app.Get("/login-cookie", func(ctx iris.Context) {
		ctx.SetCookie(&iris.Cookie{
			Name:     "SESSION_ID",
			Value:    "cookie-session-123",
			HttpOnly: true,
		})
		ctx.JSON(iris.Map{
			"message": "Logged in with cookie!",
			"method":  "cookie",
			"cookie":  "SESSION_ID=cookie-session-123",
		})
	})

	// Login as admin with cookie (Authentication Success - Cookie Admin)
	app.Get("/login-cookie-admin", func(ctx iris.Context) {
		ctx.SetCookie(&iris.Cookie{
			Name:     "SESSION_ID",
			Value:    "admin-cookie-789",
			HttpOnly: true,
		})
		ctx.JSON(iris.Map{
			"message": "Logged in as admin with cookie!",
			"method":  "cookie",
			"cookie":  "SESSION_ID=admin-cookie-789",
		})
	})

	// ========================================
	// AUTHENTICATION SUCCESS - BEARER
	// ========================================

	// Login with bearer token (Authentication Success - Bearer)
	app.Get("/login-bearer", func(ctx iris.Context) {
		ctx.JSON(iris.Map{
			"token":    "bearer-token-456",
			"username": "beareruser",
			"method":   "bearer",
			"message":  "Use this token in Authorization header",
		})
	})

	// Login as admin with bearer (Authentication Success - Bearer Admin)
	app.Get("/login-bearer-admin", func(ctx iris.Context) {
		ctx.JSON(iris.Map{
			"token":    "admin-bearer-000",
			"username": "adminbearer",
			"method":   "bearer",
			"message":  "Use this admin token in Authorization header",
		})
	})

	// ========================================
	// AUTHENTICATION FAILURE CASES
	// ========================================

	// Test with no authentication (Authentication Failure - No Auth)
	app.Get("/auth/none", func(ctx iris.Context) {
		if u, ok := compositeAuth.Authenticate(ctx); ok && u != nil {
			ctx.JSON(iris.Map{
				"authenticated": true,
				"username":      u.GetUsername(),
				"method":        "unknown",
			})
		} else {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{
				"authenticated": false,
				"error":         "No authentication provided",
				"message":       "Authentication failed - no cookie or bearer token",
			})
		}
	})

	// Test with invalid cookie (Authentication Failure - Invalid Cookie)
	app.Get("/auth/invalid-cookie", func(ctx iris.Context) {
		ctx.SetCookie(&iris.Cookie{
			Name:  "SESSION_ID",
			Value: "invalid-cookie-id",
		})
		ctx.JSON(iris.Map{
			"message": "Cookie set to invalid value. Now test /api/data to see failure.",
			"cookie":  "SESSION_ID=invalid-cookie-id",
		})
	})

	// Test with invalid bearer token (Authentication Failure - Invalid Bearer)
	app.Get("/auth/invalid-bearer", func(ctx iris.Context) {
		ctx.JSON(iris.Map{
			"message": "Now test with invalid bearer token to see failure.",
			"test":    "curl -H 'Authorization: Bearer invalid-token' <http://localhost:8083/api/data>",
		})
	})

	// ========================================
	// PROTECTED ROUTES - ACCESS GRANTED
	// ========================================

	// Protected route - Access Granted (Valid Cookie or Bearer)
	app.Get("/api/data", compositeAuth.Middleware(), func(ctx iris.Context) {
		if u, ok := compositeAuth.Authenticate(ctx); ok && u != nil {
			ctx.JSON(iris.Map{
				"message":     "You are authenticated! Access granted.",
				"username":    u.GetUsername(),
				"roles":       u.GetRoles(),
				"method":      "Works with both cookie and bearer token",
				"auth_method": getAuthMethod(ctx),
			})
		}
	})

	// Admin-only route - Access Granted (Valid Admin Cookie or Bearer)
	app.Get("/api/admin", compositeAuth.Middleware(), func(ctx iris.Context) {
		if u, ok := compositeAuth.Authenticate(ctx); ok && u != nil {
			if u.HasRole("admin") {
				ctx.JSON(iris.Map{
					"message":     "Welcome Admin! Access granted.",
					"username":    u.GetUsername(),
					"roles":       u.GetRoles(),
					"auth_method": getAuthMethod(ctx),
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

	// Manual authentication check endpoint
	app.Get("/check", func(ctx iris.Context) {
		if u, ok := compositeAuth.Authenticate(ctx); ok && u != nil {
			ctx.JSON(iris.Map{
				"authenticated": true,
				"username":      u.GetUsername(),
				"roles":         u.GetRoles(),
				"auth_method":   getAuthMethod(ctx),
			})
		} else {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{
				"authenticated": false,
				"error":         "No valid authentication found",
			})
		}
	})

	// ========================================
	// PROTECTED ROUTES - ACCESS DENIED
	// ========================================
	// These are automatically handled by Middleware() returning 401
	// Test cases:
	// - No auth: curl <http://localhost:8083/api/data>
	// - Invalid cookie: curl --cookie 'SESSION_ID=invalid' <http://localhost:8083/api/data>
	// - Invalid bearer: curl -H 'Authorization: Bearer invalid' <http://localhost:8083/api/data>

	// ========================================
	// HELPER FUNCTIONS
	// ========================================
	log.Println("==========================================")
	log.Println("Composite Authentication Example - Running on :8083")
	log.Println("==========================================")
	log.Println("")
	log.Println("AUTHENTICATION SUCCESS - COOKIE:")
	log.Println("  curl <http://localhost:8083/login-cookie>")
	log.Println("  curl --cookie 'SESSION_ID=cookie-session-123' <http://localhost:8083/api/data>")
	log.Println("")
	log.Println("AUTHENTICATION SUCCESS - BEARER:")
	log.Println("  curl <http://localhost:8083/login-bearer>")
	log.Println("  curl -H 'Authorization: Bearer bearer-token-456' <http://localhost:8083/api/data>")
	log.Println("")
	log.Println("AUTHENTICATION FAILURE - NO AUTH:")
	log.Println("  curl <http://localhost:8083/auth/none>")
	log.Println("  curl <http://localhost:8083/api/data>")
	log.Println("")
	log.Println("AUTHENTICATION FAILURE - INVALID COOKIE:")
	log.Println("  curl <http://localhost:8083/auth/invalid-cookie>")
	log.Println("  curl --cookie 'SESSION_ID=invalid-cookie-id' <http://localhost:8083/api/data>")
	log.Println("")
	log.Println("AUTHENTICATION FAILURE - INVALID BEARER:")
	log.Println("  curl -H 'Authorization: Bearer invalid-token' <http://localhost:8083/api/data>")
	log.Println("")
	log.Println("PROTECTED ROUTE - ACCESS GRANTED:")
	log.Println("  curl --cookie 'SESSION_ID=cookie-session-123' <http://localhost:8083/api/data>")
	log.Println("  curl -H 'Authorization: Bearer bearer-token-456' <http://localhost:8083/api/data>")
	log.Println("")
	log.Println("PROTECTED ROUTE - ACCESS DENIED:")
	log.Println("  curl <http://localhost:8083/api/data> (no auth - 401)")
	log.Println("  curl --cookie 'SESSION_ID=invalid' <http://localhost:8083/api/data> (invalid - 401)")
	log.Println("")
	log.Println("ADMIN ROUTE - ACCESS GRANTED:")
	log.Println("  curl --cookie 'SESSION_ID=admin-cookie-789' <http://localhost:8083/api/admin>")
	log.Println("  curl -H 'Authorization: Bearer admin-bearer-000' <http://localhost:8083/api/admin>")
	log.Println("")
	log.Println("ADMIN ROUTE - ACCESS DENIED (Forbidden):")
	log.Println("  curl --cookie 'SESSION_ID=cookie-session-123' <http://localhost:8083/api/admin>")
	log.Println("==========================================")

	app.Listen(":8083")
}

// getAuthMethod returns the authentication method used
func getAuthMethod(ctx iris.Context) string {
	if ctx.GetCookie("SESSION_ID") != "" {
		return "cookie"
	}
	if ctx.GetHeader("Authorization") != "" {
		return "bearer"
	}
	return "unknown"
}

// Ensure user.Interface is used
var _ user.Interface = (*User)(nil)
