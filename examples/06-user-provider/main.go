package main

import (
	"fmt"
	"log"

	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/user"
)

// MockUserProvider simulates a database user provider
type MockUserProvider struct{}

func (p *MockUserProvider) LoadByUsername(username string) (user.Interface, error) {
	// Simulate database lookup
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
	// Simulate database lookup by ID
	if id == "1" {
		return user.Claims{
			"sub":      "1",
			"username": "john",
			"roles":    []string{"user"},
		}, nil
	}
	return nil, fmt.Errorf("user not found")
}

// Example 6: User Provider
func main() {
	app := iris.New()

	provider := &MockUserProvider{}

	// Simulate login
	app.Post("/login", func(ctx iris.Context) {
		username := ctx.PostValue("username")

		user, err := provider.LoadByUsername(username)
		if err != nil {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{"error": "Invalid credentials"})
			return
		}

		ctx.JSON(iris.Map{
			"user_id":  user.GetID(),
			"username": user.GetUsername(),
			"roles":    user.GetRoles(),
		})
	})

	// Get user by ID
	app.Get("/user/{id}", func(ctx iris.Context) {
		id := ctx.Params().Get("id")

		user, err := provider.LoadByID(id)
		if err != nil {
			ctx.StatusCode(404)
			ctx.JSON(iris.Map{"error": "User not found"})
			return
		}

		ctx.JSON(iris.Map{
			"user_id":  user.GetID(),
			"username": user.GetUsername(),
			"roles":    user.GetRoles(),
		})
	})

	// Check role
	app.Get("/check-role/{username}/{role}", func(ctx iris.Context) {
		username := ctx.Params().Get("username")
		role := ctx.Params().Get("role")

		user, err := provider.LoadByUsername(username)
		if err != nil {
			ctx.StatusCode(404)
			ctx.JSON(iris.Map{"error": "User not found"})
			return
		}

		ctx.JSON(iris.Map{
			"username": username,
			"role":     role,
			"has_role": user.HasRole(role),
		})
	})

	log.Println("Server starting on :8085")
	log.Println("Test: curl -X POST <http://localhost:8085/login> -d 'username=john'")
	log.Println("Test: curl <http://localhost:8085/user/1>")
	log.Println("Test: curl <http://localhost:8085/check-role/john/editor>")
	app.Listen(":8085")
}
