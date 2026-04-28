package main

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/auth"
	"github.com/zeroSal/went-web/security"
	"github.com/zeroSal/went-web/user"
)

// MockUserProvider for this example
type MockUserProvider struct{}

func (p *MockUserProvider) LoadByUsername(username string) (user.Interface, error) {
	users := map[string]user.Claims{
		"john": {
			"sub":      "1",
			"username": "john",
			"roles":    []string{"user"},
		},
		"admin": {
			"sub":      "2",
			"username": "admin",
			"roles":    []string{"admin", "user"},
		},
	}

	if claims, ok := users[username]; ok {
		return claims, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (p *MockUserProvider) LoadByID(id any) (user.Interface, error) {
	if id == "1" {
		return user.Claims{
			"sub":      "1",
			"username": "john",
			"roles":    []string{"user"},
		}, nil
	}
	return nil, fmt.Errorf("user not found")
}

// Example 9: Complete Integration
func main() {
	app := iris.New()

	// Load security config from file
	sec, err := security.NewSecurity("config/security.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Set user provider
	sec.SetUserProvider(&MockUserProvider{})

	// Set role checker
	sec.SetRoleChecker(user.RoleCheckerFunc(
		func(ctx iris.Context, u user.Interface, role string) bool {
			return u.HasRole(role)
		},
	))

	// Login endpoint
	app.Post("/login", func(ctx iris.Context) {
		username := ctx.PostValue("username")
		password := ctx.PostValue("password")

		// Simple check
		if (username == "john" || username == "admin") && password == "secret" {
			// Get user
			u, _ := sec.UserProvider().LoadByUsername(username)

			// Generate JWT
			jwtAuth := sec.Authenticator().(*auth.JWT)
			claims := jwt.MapClaims{
				"sub":      u.GetID(),
				"username": u.GetUsername(),
				"roles":    u.GetRoles(),
			}

			token, err := jwtAuth.GenerateToken(claims)
			if err != nil {
				ctx.StatusCode(500)
				ctx.JSON(iris.Map{"error": "Failed to generate token"})
				return
			}

			ctx.JSON(iris.Map{
				"token":    token,
				"username": u.GetUsername(),
				"roles":    u.GetRoles(),
			})
		} else {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{"error": "Invalid credentials"})
		}
	})

	// Public endpoint
	app.Get("/public/info", func(ctx iris.Context) {
		ctx.JSON(iris.Map{
			"message": "This is public information",
			"time":    time.Now(),
		})
	})

	// Protected endpoint
	app.Get("/api/profile", sec.Middleware(), func(ctx iris.Context) {
		// Get user from JWT claims
		if claims, ok := sec.Authenticator().(*auth.JWT).Authenticate(ctx); ok {
			mapClaims := claims.(jwt.MapClaims)
			ctx.JSON(iris.Map{
				"user_id":  mapClaims["sub"],
				"username": mapClaims["username"],
			})
		}
	})

	// Admin only endpoint
	app.Get("/admin/users", sec.Middleware(), func(ctx iris.Context) {
		// Check if user has admin role
		if claims, ok := sec.Authenticator().(*auth.JWT).Authenticate(ctx); ok {
			userClaims := claims.(jwt.MapClaims)
			roles := userClaims["roles"].([]interface{})

			hasAdmin := false
			for _, r := range roles {
				if r == "admin" {
					hasAdmin = true
					break
				}
			}

			if hasAdmin {
				ctx.JSON(iris.Map{
					"users": []string{"john", "admin"},
					"count": 2,
				})
			} else {
				ctx.StatusCode(403)
				ctx.JSON(iris.Map{"error": "Forbidden: Admin access required"})
			}
		}
	})

	log.Println("Server starting on :8089")
	log.Println("Login: curl -X POST <http://localhost:8089/login> -d 'username=admin&password=secret'")
	log.Println("Profile: curl -H 'Authorization: Bearer <token>' <http://localhost:8089/api/profile>")
	log.Println("Admin: curl -H 'Authorization: Bearer <token>' <http://localhost:8089/admin/users>")
	app.Listen(":8089")
}
