# User Management

## Overview

The `user` package provides interfaces for managing users, roles, and permissions.

## Claims

`user.Claims` implements `user.Interface` and represents authenticated user data.

```go
import "github.com/zeroSal/went-web/user"

// Create claims with user data
claims := user.Claims{
    "sub":      "user123",      // User ID (required for GetID())
    "username": "john",         // Username
    "roles":    []string{"admin", "user"}, // Roles array
    "email":    "john@example.com",
}

// Get user ID
id := claims.GetID()           // Returns: "user123"

// Get username
username := claims.GetUsername() // Returns: "john" (or "user123" if not set)

// Get roles
roles := claims.GetRoles()      // Returns: []string{"admin", "user"}

// Check specific role
if claims.HasRole("admin") {
    fmt.Println("User is admin")
}

// Use in JWT claims
jwtClaims := jwt.MapClaims(claims) // Convert to jwt.MapClaims
```

**Claims Methods:**
- `GetID() any` - Returns user ID (from `"sub"` key)
- `GetUsername() string` - Returns username (from `"username"` or `"sub"`)
- `GetRoles() []string` - Returns roles (from `"roles"` array)
- `HasRole(role string) bool` - Checks if user has a specific role

## User Interface

Define a user object with `user.Interface`:

```go
import "github.com/zeroSal/went-web/user"

type Interface interface {
    GetID() any
    GetUsername() string
    GetRoles() []string
    HasRole(role string) bool
}
```

## User Provider

Load users from a data source (database, API, etc.):

```go
import (
    "github.com/zeroSal/went-web/user"
    "github.com/zeroSal/went-web/security"
)

// Implement user.Provider
type DBUserProvider struct {
    // database connection, etc.
}

func (p *DBUserProvider) LoadByUsername(username string) (user.Interface, error) {
    // Load user from database by username
    return user.Claims{
        "sub":      "1",
        "username": username,
        "roles":    []string{"user"},
    }, nil
}

func (p *DBUserProvider) LoadByID(id any) (user.Interface, error) {
    // Load user by ID from database
    return user.Claims{
        "sub":      id,
        "username": "john",
        "roles":    []string{"user"},
    }, nil
}

// Register with Security
func main() {
    app := iris.New()
    sec, _ := security.NewSecurity("security.yaml")

    // Set user provider
    sec.SetUserProvider(&DBUserProvider{})

    app.Use(sec.Middleware())
    app.Listen(":8080")
}
```

## Role Checker

Custom role checking logic:

```go
import (
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/user"
    "github.com/zeroSal/went-web/security"
)

// Implement user.RoleChecker
type MyRoleChecker struct{}

func (rc *MyRoleChecker) CheckRole(ctx iris.Context, u user.Interface, role string) bool {
    // Custom role checking logic
    // Check against database, external service, etc.
    return u.HasRole(role)
}

// Or use RoleCheckerFunc for simple cases
roleChecker := user.RoleCheckerFunc(
    func(ctx iris.Context, u user.Interface, role string) bool {
        return u.HasRole(role)
    },
)

// Register with Security
func main() {
    app := iris.New()
    sec, _ := security.NewSecurity("security.yaml")

    // Set role checker
    sec.SetRoleChecker(roleChecker)

    app.Use(sec.Middleware())
    app.Listen(":8080")
}
```

## Complete Example

```go
package main

import (
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/security"
    "github.com/zeroSal/went-web/user"
)

// User provider that loads from "database"
type DBUserProvider struct{}

func (p *DBUserProvider) LoadByUsername(username string) (user.Interface, error) {
    // In real app, query database here
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

func (p *DBUserProvider) LoadByID(id any) (user.Interface, error) {
    // Load by ID from database
    return user.Claims{
        "sub":      id,
        "username": "john",
        "roles":    []string{"user"},
    }, nil
}

func main() {
    app := iris.New()
    sec, _ := security.NewSecurity("security.yaml")

    // Set user provider
    sec.SetUserProvider(&DBUserProvider{})

    // Set role checker
    sec.SetRoleChecker(user.RoleCheckerFunc(
        func(ctx iris.Context, u user.Interface, role string) bool {
            return u.HasRole(role)
        },
    ))

    // Apply security middleware
    app.Use(sec.Middleware())

    app.Listen(":8080")
}
```
