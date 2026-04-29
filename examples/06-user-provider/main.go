package main

import (
	"fmt"
	"log"
	"slices"

	"github.com/kataras/iris/v12"
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

// MockUserProvider simulates a database user provider
type MockUserProvider struct{}

func (p *MockUserProvider) Load(credential any) (user.Interface, error) {
	// credential can be username (string) or user ID
	switch c := credential.(type) {
	case string:
		// Check if it's a username
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
		if u, ok := users[c]; ok {
			return u, nil
		}

		// Check if it's a session token
		tokens := map[string]*User{
			"session-john-123": {
				ID:       "1",
				Username: "john",
				Roles:    []string{"user"},
			},
			"session-admin-456": {
				ID:       "2",
				Username: "admin",
				Roles:    []string{"admin", "user"},
			},
		}
		if u, ok := tokens[c]; ok {
			return u, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

// Example 6: User Provider - Complete Test Cases
func main() {
	app := iris.New()

	provider := &MockUserProvider{}

	// ========================================
	// AUTHENTICATION SUCCESS
	// ========================================

	// Login with username (Authentication Success)
	app.Get("/login", func(ctx iris.Context) {
		ctx.JSON(iris.Map{
			"user_id":  "1",
			"username": "john",
			"roles":    []string{"user"},
			"message":  "Login successful",
		})
	})

	app.Post("/login", func(ctx iris.Context) {
		username := ctx.PostValue("username")

		// Simple lookup by username (test sends only username)
		u, err := provider.Load(username)
		if err != nil {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{"error": "User not found"})
			return
		}

		ctx.JSON(iris.Map{
			"user_id":  u.GetID(),
			"username": u.GetUsername(),
			"roles":    u.GetRoles(),
			"message":  "Login successful",
		})
	})

	// Login with GET for easier testing (Authentication Success)
	app.Get("/login/{username}", func(ctx iris.Context) {
		username := ctx.Params().Get("username")

		u, err := provider.Load(username)
		if err != nil {
			ctx.StatusCode(404)
			ctx.JSON(iris.Map{"error": "User not found"})
			return
		}

		ctx.JSON(iris.Map{
			"user_id":  u.GetID(),
			"username": u.GetUsername(),
			"roles":    u.GetRoles(),
			"message":  "Login successful (GET)",
		})
	})

	// ========================================
	// AUTHENTICATION FAILURE
	// ========================================

	// Test with non-existent user (Authentication Failure)
	app.Get("/login/invalid", func(ctx iris.Context) {
		u, err := provider.Load("nonexistent")
		if err != nil {
			ctx.StatusCode(404)
			ctx.JSON(iris.Map{
				"error":   "User not found",
				"message": "Authentication failed - invalid username",
			})
			return
		}
		ctx.JSON(iris.Map{"user": u})
	})

	// ========================================
	// USER LOOKUP
	// ========================================

	// Get user by ID
	app.Get("/user/{id}", func(ctx iris.Context) {
		id := ctx.Params().Get("id")

		u, err := provider.Load(id)
		if err != nil {
			ctx.StatusCode(404)
			ctx.JSON(iris.Map{"error": "User not found"})
			return
		}

		ctx.JSON(iris.Map{
			"user_id":  u.GetID(),
			"username": u.GetUsername(),
			"roles":    u.GetRoles(),
		})
	})

	// Check role
	app.Get("/check-role/{username}/{role}", func(ctx iris.Context) {
		username := ctx.Params().Get("username")
		role := ctx.Params().Get("role")

		u, err := provider.Load(username)
		if err != nil {
			ctx.StatusCode(404)
			ctx.JSON(iris.Map{"error": "User not found"})
			return
		}

		ctx.JSON(iris.Map{
			"username": username,
			"role":     role,
			"has_role": u.HasRole(role),
		})
	})

	// ========================================
	// SESSION TOKEN TESTS
	// ========================================

	// Simulate session token lookup (Authentication Success)
	app.Get("/session/{token}", func(ctx iris.Context) {
		token := ctx.Params().Get("token")

		u, err := provider.Load(token)
		if err != nil {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{
				"error":   "Invalid session token",
				"message": "Authentication failed",
			})
			return
		}

		ctx.JSON(iris.Map{
			"user_id":  u.GetID(),
			"username": u.GetUsername(),
			"roles":    u.GetRoles(),
			"message":  "Session valid",
		})
	})

	// ========================================
	// TEST INSTRUCTIONS
	// ========================================
	log.Println("==========================================")
	log.Println("User Provider Example - Running on :8085")
	log.Println("==========================================")
	log.Println("")
	log.Println("AUTHENTICATION SUCCESS:")
	log.Println("  curl -X POST -d 'username=john&password=secret' <http://localhost:8085/login>")
	log.Println("  curl <http://localhost:8085/login/john>")
	log.Println("")
	log.Println("AUTHENTICATION FAILURE:")
	log.Println("  curl <http://localhost:8085/login/nonexistent>")
	log.Println("  curl <http://localhost:8085/login/invalid>")
	log.Println("")
	log.Println("USER LOOKUP:")
	log.Println("  curl <http://localhost:8085/user/1>")
	log.Println("  curl <http://localhost:8085/user/2>")
	log.Println("")
	log.Println("ROLE CHECK:")
	log.Println("  curl <http://localhost:8085/check-role/john/admin>")
	log.Println("  curl <http://localhost:8085/check-role/admin/admin>")
	log.Println("")
	log.Println("SESSION TOKEN TEST:")
	log.Println("  curl <http://localhost:8085/session/session-john-123>")
	log.Println("  curl <http://localhost:8085/session/session-admin-456>")
	log.Println("  curl <http://localhost:8085/session/invalid-token>")
	log.Println("==========================================")

	app.Listen(":8085")
}

// Ensure interfaces are used
var _ session.ProviderInterface = (*MockUserProvider)(nil)
var _ user.Interface = (*User)(nil)
