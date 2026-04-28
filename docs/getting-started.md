# Getting Started

## Install

```bash
go get github.com/zeroSal/went-web
```

Dependencies (auto-installed): `iris/v12`, `golang-jwt/jwt/v5`, `yaml.v3`

## Quick Start

```go
package main

import (
    "embed"
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/security"
)

//go:embed security.yaml routes.yaml
var embedFS embed.FS

func main() {
    app := iris.New()

    // Load security config from embedded files
    sec, err := security.NewSecurityFromEmbed(embedFS, "security.yaml")
    if err != nil {
        panic(err)
    }

    // Load routes config
    routes, err := security.LoadRoutesConfigFromEmbed(embedFS, "routes.yaml")
    if err != nil {
        panic(err)
    }
    sec.SetRoutes(routes)

    // Register handlers
    sec.RegisterHandler("Home", "GET", func(ctx iris.Context) {
        ctx.WriteString("Hello World") // Simple response
    })

    // Apply security middleware (checks auth on every request)
    app.Use(sec.Middleware())

    // Register all routes defined in routes.yaml
    sec.RegisterRoutes(app)

    app.Listen(":8080")
}
```

## YAML Config

**security.yaml:**
```yaml
firewalls:
  main:
    pattern: "^/"                    # Regex for protected paths
    auth:
      jwt:
        secret: "your-32-byte-secret-key-here!!!!"
        expiry: 3600

access:
  - path: /login
    require: IS_AUTHENTICATED_ANONYMOUSLY  # Public
  - path: /admin
    require: ROLE_ADMIN                    # Admin only
  - path: /*
    require: AUTH_REQUIRED                 # Any authenticated user
```

**routes.yaml:**
```yaml
routes:
  - path: /
    method: [GET]
    handler: Home                    # Calls registered "Home" GET handler
    require: AUTH_REQUIRED

  - path: /api/data
    handlers:
      GET: Api.GetData              # Controller.Method format
      POST: Api.PostData
    require: AUTH_REQUIRED
```
