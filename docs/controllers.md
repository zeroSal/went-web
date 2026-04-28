# Controller System

## Overview

Controllers organize HTTP handlers in a Symfony-style architecture using reflection for automatic method discovery.

## Creating Controllers

Implement `controller.Interface` and define routes:

```go
import (
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/controller"
)

// Embed controller.Base for convenience
type HomeController struct {
    controller.Base
}

// Return route configuration
func (c *HomeController) GetConfiguration() controller.Configuration {
    return controller.Configuration{
        Name: "Home", // Controller name for handler registration
        Routes: []controller.Route{
            {Path: "/", Method: "GET", MethodName: "Index", Handler: c.Index},
            {Path: "/about", Method: "GET", MethodName: "About", Handler: c.About},
            {Path: "/contact", Method: "GET", MethodName: "Contact", Handler: c.Contact},
        },
    }
}

// Handler methods
func (c *HomeController) Index(ctx iris.Context) {
    ctx.WriteString("Welcome to Home Page!")
}

func (c *HomeController) About(ctx iris.Context) {
    ctx.WriteString("About Us Page")
}

func (c *HomeController) Contact(ctx iris.Context) {
    ctx.WriteString("Contact Page")
}
```

## Controller Registry

The registry uses reflection to discover handler methods automatically:

```go
import (
    "github.com/zeroSal/went-web/controller"
    "github.com/zeroSal/went-web/security"
)

func main() {
    app := iris.New()
    sec, _ := security.NewSecurity("security.yaml")

    // Create registry
    registry := controller.NewRegistry()

    // Register controllers
    registry.Register(&HomeController{})
    registry.Register(&APIController{})

    // Get all handlers for security registration
    for _, h := range registry.Handlers() {
        // h.Controller = "Home", h.Method = "GET", h.Handler = handler
        sec.RegisterHandler(h.Controller, h.Method, h.Handler)
    }

    // Load and set routes
    routes, _ := security.LoadRoutesConfig("routes.yaml")
    sec.SetRoutes(routes)

    // Apply middleware
    app.Use(sec.Middleware())

    // Register routes
    sec.RegisterRoutes(app)

    app.Listen(":8080")
}
```

**Automatic Handler Discovery:**
The registry reflects on controllers and finds methods that:
- Have exactly one parameter (the receiver)
- The parameter type is `iris.Context`

## Route Configuration

Define routes in YAML or programmatically:

```yaml
# routes.yaml
routes:
  - path: /
    method: [GET]
    handler: Home.Index              # ControllerName.MethodName
    require: AUTH_REQUIRED

  - path: /api/data
    handlers:
      GET: Api.GetData              # Calls APIController.GetData
      POST: Api.PostData            # Calls APIController.PostData
      PUT: Api.UpdateData
      DELETE: Api.DeleteData

  - path: /users/{id}
    method: [GET]
    handler: User.Show              # Wildcard path support
```

**Handler Format:** `ControllerName.MethodName` (e.g., `Home.Index`, `Api.GetData`)

## Controller Types

```go
// Interface that controllers must implement
type Interface interface {
    GetConfiguration() Configuration
}

// Configuration holds controller routes
type Configuration struct {
    Name   string          // Controller name
    Routes []Route        // List of routes
}

// Route defines a single endpoint
type Route struct {
    Path       string
    Method     string
    MethodName string      // For reflection-based discovery
    Handler    iris.Handler
}

// Handler for registration with security
type Handler struct {
    Controller string
    Method     string
    Handler    iris.Handler
}
```

## Using with Security

```go
import (
    "github.com/zeroSal/went-web/security"
    "github.com/zeroSal/went-web/controller"
)

func main() {
    app := iris.New()
    sec, _ := security.NewSecurity("security.yaml")

    // Create registry and register controllers
    registry := controller.NewRegistry()
    registry.Register(&HomeController{})
    registry.Register(&AdminController{})

    // Register handlers with security
    for _, h := range registry.Handlers() {
        sec.RegisterHandler(h.Controller, h.Method, h.Handler)
    }

    // Load and set routes
    routes, _ := security.LoadRoutesConfig("routes.yaml")
    sec.SetRoutes(routes)

    // Apply security middleware
    app.Use(sec.Middleware())

    // Register routes
    sec.RegisterRoutes(app)

    app.Listen(":8080")
}
```

## Complete Example

```go
package main

import (
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/controller"
    "github.com/zeroSal/went-web/security"
)

// User controller with multiple endpoints
type UserController struct {
    controller.Base
}

func (c *UserController) GetConfiguration() controller.Configuration {
    return controller.Configuration{
        Name: "User",
        Routes: []controller.Route{
            {Path: "/users", Method: "GET", MethodName: "List", Handler: c.List},
            {Path: "/users/{id}", Method: "GET", MethodName: "Show", Handler: c.Show},
            {Path: "/users", Method: "POST", MethodName: "Create", Handler: c.Create},
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

func (c *UserController) Create(ctx iris.Context) {
    var data map[string]interface{}
    if err := ctx.ReadJSON(&data); err != nil {
        ctx.StatusCode(400)
        return
    }
    ctx.JSON(iris.Map{"status": "created"})
}

func main() {
    app := iris.New()
    sec, _ := security.NewSecurity("security.yaml")

    // Register controllers
    registry := controller.NewRegistry()
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
