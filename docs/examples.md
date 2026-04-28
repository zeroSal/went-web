# Complete Examples

## Example 1: JWT Authentication

**security.yaml:**
```yaml
firewalls:
  main:
    pattern: "^/"
    auth:
      jwt:
        secret: "my-super-secret-key-32-bytes-long!!"
        expiry: 86400

access:
  - path: /login
    require: IS_AUTHENTICATED_ANONYMOUSLY
  - path: /api/*
    require: AUTH_REQUIRED
```

**main.go:**
```go
package main

import (
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/security"
    "github.com/zeroSal/went-web/auth"
)

func main() {
    app := iris.New()
    sec, _ := security.NewSecurity("security.yaml")

    // Login handler - generates JWT token
    app.Post("/login", func(ctx iris.Context) {
        username := ctx.PostValue("username")
        password := ctx.PostValue("password")

        // Validate credentials (your logic here)
        if username == "admin" && password == "secret" {
            // Get JWT authenticator from security
            jwtAuth := sec.Authenticator().(*auth.JWT)

            // Create claims with user info
            claims := jwt.MapClaims{
                "sub":      "1",
                "username": username,
                "roles":    []string{"admin"},
            }

            // Generate token
            token, _ := jwtAuth.GenerateToken(claims)
            ctx.JSON(iris.Map{"token": token})
        } else {
            ctx.StatusCode(401)
            ctx.JSON(iris.Map{"error": "Invalid credentials"})
        }
    })

    // Protected route - requires valid JWT
    app.Get("/api/data", sec.Middleware(), func(ctx iris.Context) {
        // Get claims from context (set by middleware)
        if claims, ok := sec.Authenticator().(*auth.JWT).Authenticate(ctx); ok {
            ctx.JSON(iris.Map{
                "message": "Protected data",
                "user_id": claims.(jwt.MapClaims)["sub"],
            })
        }
    })

    app.Listen(":8080")
}
```

## Example 2: Multi-Auth Strategy

**security.yaml:**
```yaml
firewalls:
  main:
    pattern: "^/"
    auth:
      cookie:
        name: "SESSION_ID"
      bearer:
        enabled: true
      jwt:
        secret: "secret-key-here!!!!"
        expiry: 3600

access:
  - path: /public/*
    require: IS_AUTHENTICATED_ANONYMOUSLY
  - path: /admin/*
    require: ROLE_ADMIN
  - path: /*
    require: AUTH_REQUIRED
```

**main.go:**
```go
package main

import (
    "embed"
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/security"
)

//go:embed security.yaml routes.yaml
var fs embed.FS

func main() {
    app := iris.New()
    sec, _ := security.NewSecurityFromEmbed(fs, "security.yaml")

    routes, _ := security.LoadRoutesConfigFromEmbed(fs, "routes.yaml")
    sec.SetRoutes(routes)

    // Register handlers
    sec.RegisterHandler("Home", "GET", func(ctx iris.Context) {
        ctx.WriteString("Home")
    })
    sec.RegisterHandler("Admin", "GET", func(ctx iris.Context) {
        ctx.WriteString("Admin Panel")
    })

    // Apply security (tries Cookie, Bearer, JWT in order)
    app.Use(sec.Middleware())

    // Register routes
    sec.RegisterRoutes(app)

    app.Listen(":8080")
}
```

## Example 3: User Provider + Role Checker

```go
package main

import (
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/security"
    "github.com/zeroSal/went-web/user"
)

// Mock user provider
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
    sec.SetUserProvider(&MockUserProvider{})

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

## Example 4: Full App with Controllers

**security.yaml:**
```yaml
firewalls:
  main:
    pattern: "^/"
    auth:
      jwt:
        secret: "your-32-byte-secret-key-here!!!!"
        expiry: 3600

access:
  - path: /login
    require: IS_AUTHENTICATED_ANONYMOUSLY
  - path: /api/*
    require: AUTH_REQUIRED
```

**routes.yaml:**
```yaml
routes:
  - path: /login
    method: [GET, POST]
    handler: Auth.login
    require: IS_AUTHENTICATED_ANONYMOUSLY

  - path: /api/users
    method: [GET]
    handler: User.list
    require: AUTH_REQUIRED

  - path: /api/users/{id}
    method: [GET]
    handler: User.show
    require: AUTH_REQUIRED
```

**main.go:**
```go
package main

import (
    "embed"
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/security"
    "github.com/zeroSal/went-web/controller"
)

//go:embed security.yaml routes.yaml
var fs embed.FS

// Auth controller
type AuthController struct {
    controller.Base
}

func (c *AuthController) GetConfiguration() controller.Configuration {
    return controller.Configuration{
        Name: "Auth",
        Routes: []controller.Route{
            {Path: "/login", Method: "GET", MethodName: "Login", Handler: c.Login},
            {Path: "/login", Method: "POST", MethodName: "LoginPost", Handler: c.LoginPost},
        },
    }
}

func (c *AuthController) Login(ctx iris.Context) {
    ctx.WriteString("Login Page")
}

func (c *AuthController) LoginPost(ctx iris.Context) {
    ctx.WriteString("Login Handler")
}

// User controller
type UserController struct {
    controller.Base
}

func (c *UserController) GetConfiguration() controller.Configuration {
    return controller.Configuration{
        Name: "User",
        Routes: []controller.Route{
            {Path: "/api/users", Method: "GET", MethodName: "List", Handler: c.List},
            {Path: "/api/users/{id}", Method: "GET", MethodName: "Show", Handler: c.Show},
        },
    }
}

func (c *UserController) List(ctx iris.Context) {
    ctx.JSON(iris.Map{"users": []string{}})
}

func (c *UserController) Show(ctx iris.Context) {
    id := ctx.Params().Get("id")
    ctx.JSON(iris.Map{"id": id})
}

func main() {
    app := iris.New()

    // Load security config
    sec, _ := security.NewSecurityFromEmbed(fs, "security.yaml")

    // Load routes config
    routes, _ := security.LoadRoutesConfigFromEmbed(fs, "routes.yaml")
    sec.SetRoutes(routes)

    // Register controllers
    registry := controller.NewRegistry()
    registry.Register(&AuthController{})
    registry.Register(&UserController{})

    // Register handlers with security
    for _, h := range registry.Handlers() {
        sec.RegisterHandler(h.Controller, h.Method, h.Handler)
    }

    // Apply security middleware
    app.Use(sec.Middleware())

    // Register routes
    sec.RegisterRoutes(app)

    app.Listen(":8080")
}
```
