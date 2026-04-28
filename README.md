# went-web

Symfony-style security and controller library for Go with Iris web framework.

## Features

- **Authentication**: Cookie, Bearer, JWT, Composite strategies
- **Authorization**: Role-based access control (RBAC)
- **Security**: Firewalls, access rules, CSRF protection
- **Session Management**: Iris sessions with security integration
- **Controllers**: Symfony-style controller system with reflection
- **YAML Config**: Symfony-like `security.yaml` and `routes.yaml`

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
    routes, _ := security.LoadRoutesConfigFromEmbed(embedFS, "routes.yaml")
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

## YAML Configuration

**security.yaml:**
```yaml
firewalls:
  main:
    pattern: "^/"              # Regex pattern for paths
    auth:
      jwt:
        secret: "your-32-byte-secret-key-here!!!!"
        expiry: 3600

access:
  - path: /login
    require: IS_AUTHENTICATED_ANONYMOUSLY  # Public
  - path: /admin
    require: ROLE_ADMIN              # Admin only
  - path: /*
    require: AUTH_REQUIRED           # Any authenticated user
```

**routes.yaml:**
```yaml
routes:
  - path: /
    method: [GET]
    handler: Home              # Calls registered "Home" GET handler
    require: AUTH_REQUIRED

  - path: /api/users
    method: [GET, POST]
    handler: User              # Auto-discovers GetData/PostData methods
```

## Installation

```bash
go get github.com/zeroSal/went-web
```

## Testing

```bash
make test           # Unit tests with coverage
make test-all       # Unit + functional tests (all examples)
make test-coverage  # Show coverage summary
make coverage-html  # Generate HTML coverage report
```

## Examples

```bash
# Run any example
cd examples/01-cookie-auth && go run main.go

# Or run all functional tests
make test-all
```

| # | Example | Port | Auth Method |
|---|---------|------|-------------|
| 1 | [Cookie Auth](examples/01-cookie-auth/) | 8080 | Cookie-based |
| 2 | [Bearer Auth](examples/02-bearer-auth/) | 8081 | Bearer token |
| 3 | [JWT Auth](examples/03-jwt-auth/) | 8082 | JWT with token generation |
| 4 | [Composite Auth](examples/04-composite-auth/) | 8083 | Cookie + Bearer |
| 5 | [User Claims](examples/05-user-claims/) | 8084 | Claims interface |
| 6 | [User Provider](examples/06-user-provider/) | 8085 | Custom user provider |
| 7 | [Controllers](examples/07-controllers/) | 8086 | Controller system |
| 8 | [Full Security](examples/08-security-full/) | 8088 | Complete YAML config |
| 9 | [Integration](examples/09-integration/) | 8089 | Full integration |
| 10 | [Routes](examples/10-routes/) | 8090 | Route configuration |
| 11 | [Session](examples/11-session/) | 8091 | Session management |
| 12 | [CSRF](examples/12-csrf/) | 8092 | CSRF protection |

## Documentation

| Topic | File |
|-------|------|
| Getting Started | [docs/getting-started.md](docs/getting-started.md) |
| Authentication | [docs/authentication.md](docs/authentication.md) |
| Security Config | [docs/security.md](docs/security.md) |
| User Management | [docs/user-management.md](docs/user-management.md) |
| Controllers | [docs/controllers.md](docs/controllers.md) |
| Examples | [docs/examples.md](docs/examples.md) |
