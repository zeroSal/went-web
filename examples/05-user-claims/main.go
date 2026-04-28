package main

import (
	"fmt"
	"log"

	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/user"
)

// Example 5: User Claims
func main() {
	app := iris.New()

	// Create user claims
	claims := user.Claims{
		"sub":      "user123",
		"username": "john_doe",
		"roles":    []string{"user", "editor"},
		"email":    "john@example.com",
	}

	// Route to demonstrate claims methods
	app.Get("/user/info", func(ctx iris.Context) {
		// Get user ID
		id := claims.GetID()
		ctx.WriteString("User ID: " + id.(string) + "\n")

		// Get username
		username := claims.GetUsername()
		ctx.WriteString("Username: " + username + "\n")

		// Get roles
		roles := claims.GetRoles()
		ctx.WriteString("Roles: " + fmt.Sprint(roles) + "\n")

		// Check specific role
		if claims.HasRole("editor") {
			ctx.WriteString("User has 'editor' role!\n")
		}
		if !claims.HasRole("admin") {
			ctx.WriteString("User does NOT have 'admin' role\n")
		}
	})

	// Test HasRole function
	app.Get("/user/has-role/{role}", func(ctx iris.Context) {
		role := ctx.Params().Get("role")
		if claims.HasRole(role) {
			ctx.JSON(iris.Map{
				"role":     role,
				"has_role": true,
			})
		} else {
			ctx.JSON(iris.Map{
				"role":     role,
				"has_role": false,
			})
		}
	})

	log.Println("Server starting on :8084")
	log.Println("Test: curl <http://localhost:8084/user/info>")
	log.Println("Test: curl <http://localhost:8084/user/has-role/editor>")
	app.Listen(":8084")
}
