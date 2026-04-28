# Security Configuration

## Overview

The `security` package handles all security: config loading, authentication, authorization, sessions, CSRF.

## Security Config (security.yaml)

```yaml
# Firewalls define authentication strategies for URL patterns
firewalls:
  main:                    # Firewall name (arbitrary)
    pattern: "^/"          # Regex pattern to match paths
    auth:
      cookie:              # Cookie-based auth
        name: "SESSION_ID"
      bearer:              # Bearer token auth
        enabled: true
      jwt:                 # JWT auth
        secret: "your-32-byte-secret-key-here!!!!"
        expiry: 3600       # Seconds

# Access rules define which paths require authentication
access:
  - path: /login
    require: IS_AUTHENTICATED_ANONYMOUSLY  # Public path
  - path: /admin
    require: ROLE_ADMIN                    # Requires admin role
  - path: /
    require: AUTH_REQUIRED                 # Any authenticated user

# Session configuration (optional)
session:
  cookie: "SESSION_ID"                     # Session cookie name
  cookie_path: "/"                         # Cookie path
  expires: 3600                            # Expiration in seconds
  secure: true                             # HTTPS only
  allow_reclaim: false                     # Allow session reclaim

# CSRF protection (optional)
csrf:
  enabled: true
  secret: "your-32-byte-auth-key-here!!!!"
  secure: true
  same_site: "Lax"                         # Lax, Strict, None
  field_name: "csrf_token"                  # Form field name
  header_name: "X-CSRF-Token"              # Header name

# Logout configuration (optional)
logout:
  enabled: true
  logout_url: "/logout"
  delete_cookies:                          # Cookies to delete on logout
    - "SESSION_ID"
    - "jwt"
  redirect_url: "/login"

# Entry point (optional - redirects unauthenticated users)
entry_point:
  login_url: "/login"
  code: 302                                # Redirect code
```

## Loading Config

```go
import "github.com/zeroSal/went-web/security"

// From file
sec, err := security.NewSecurity("security.yaml")

// From embedded filesystem
//go:embed security.yaml
var fs embed.FS
sec, err := security.NewSecurityFromEmbed(fs, "security.yaml")

// From config struct
config := &security.SecurityConfig{...}
sec, err := security.NewSecurityFromConfig(config)

// From YAML bytes
data := []byte(...)
config, err := security.LoadSecurityConfigFromBytes(data)
sec, err := security.NewSecurityFromConfig(config)
```

## Routes Config (routes.yaml)

```yaml
routes:
  - path: /
    method: [GET]
    handler: Home.index                     # Format: ControllerName.MethodName
    require: AUTH_REQUIRED

  - path: /login
    method: [GET, POST]
    handler: Auth.login
    require: IS_AUTHENTICATED_ANONYMOUSLY

  # Per-method handlers
  - path: /profile
    handlers:
      GET: Profile.show
      POST: Profile.update
    require: AUTH_REQUIRED

  # Wildcard path
  - path: /api/*
    method: [GET, POST, PUT, DELETE]
    handler: Api.handler
```

**Loading routes:**
```go
import "github.com/zeroSal/went-web/security"

// From file
routes, err := security.LoadRoutesConfig("routes.yaml")

// From embed
routes, err := security.LoadRoutesConfigFromEmbed(fs, "routes.yaml")

// Set on security
sec.SetRoutes(routes)
```

## Security Middleware

```go
sec, _ := security.NewSecurity("security.yaml")

// Get authenticator
auth := sec.Authenticator()

// Apply middleware (checks auth on every request)
app.Use(sec.Middleware())
```

## Access Rules

```yaml
access:
  - path: /login
    require: IS_AUTHENTICATED_ANONYMOUSLY  # Public - no auth required

  - path: /admin/*
    require: ROLE_ADMIN                    # Requires admin role

  - path: /api/*
    require: AUTH_REQUIRED                 # Any authenticated user

  - path: /editor/*
    require: ROLE_EDITOR                   # Requires editor role
```

**Special Requirements:**
- `IS_AUTHENTICATED_ANONYMOUSLY` - Path is public
- `AUTH_REQUIRED` - Any authenticated user
- `ROLE_*` - Specific role required (e.g., `ROLE_ADMIN`, `ROLE_USER`)

**Path Matching:**
- Wildcards: `/admin/*` matches `/admin/users`, `/admin/settings`
- Method-specific: `GET /api/*` matches only GET requests
- Placeholders: `/users/{id}` matches `/users/123`

## Session Management

```yaml
session:
  cookie: "SESSION_ID"
  cookie_path: "/"
  expires: 3600      # Seconds
  secure: true       # HTTPS only
```

**Usage in code:**
```go
sec, _ := security.NewSecurity("security.yaml")

// Get session manager
session := sec.Session()

// Use in handler
handler := func(ctx iris.Context) {
    s := session.Start(ctx)
    s.Set("user_id", "123")           // Set session value
    userID := s.GetString("user_id") // Get session value
}
```

## CSRF Protection

```yaml
csrf:
  enabled: true
  secret: "your-csrf-secret-here!!!!"
  secure: true
  same_site: "Lax"         # Lax, Strict, or None
  field_name: "csrf_token"
  header_name: "X-CSRF-Token"
```

**Usage:**
```go
sec, _ := security.NewSecurity("security.yaml")

// Get CSRF middleware
csrfMiddleware := sec.CSRF()

// Apply to specific routes
app.Post("/form", csrfMiddleware, func(ctx iris.Context) {
    // Form handling - CSRF validated automatically
})
```

## API Reference

**Constructor Functions:**
- `NewSecurity(configPath string) (*Security, error)` - Create from file
- `NewSecurityFromEmbed(efs embed.FS, path string) (*Security, error)` - From embed
- `NewSecurityFromConfig(config *SecurityConfig) (*Security, error)` - From struct

**Security Methods:**
- `Middleware() iris.Handler` - Returns main security middleware
- `Authenticator() auth.Interface` - Returns the authenticator
- `Session() *sessions.Sessions` - Returns session manager
- `CSRF() iris.Handler` - Returns CSRF middleware
- `SetUserProvider(provider user.Provider)` - Sets user provider
- `SetRoleChecker(checker user.RoleChecker)` - Sets role checker
- `SetRoutes(routes []RouteConfig)` - Sets route configurations
- `RegisterHandler(controller, method string, handler iris.Handler)` - Registers handlers
- `RegisterRoutes(app *iris.Application)` - Registers all routes to Iris app
