# went-web - Symfony-style Security & Controller Library for Go

## Table of Contents
- [Overview](#overview)
- [Installation](#installation)
- [Configuration](#configuration)
  - [security.yaml](#securityyaml)
  - [routes.yaml](#routesyaml)
- [Package `auth` - Authentication](#package-auth---authentication)
  - [Interface](#authinterface)
  - [Cookie Authentication](#cookie-authentication)
  - [Bearer Authentication](#bearer-authentication)
  - [Composite Authentication](#composite-authentication)
- [Package `user` - User Management](#package-user---user-management)
- [Package `session` - Session Provider](#package-session---session-provider)
- [Package `controller` - Controllers](#package-controller---controllers)
- [Package `security` - Core Security](#package-security---core-security)
  - [SecurityConfig](#securityconfig)
  - [Firewall & Access Control](#firewall--access-control)
  - [CSRF Protection](#csrf-protection)
  - [Session Configuration](#session-configuration)
  - [Route Registration](#route-registration)
- [Complete Examples](#complete-examples)
  - [Example 1: Cookie Authentication](#example-1-cookie-authentication)
  - [Example 2: Bearer Token Authentication](#example-2-bearer-token-authentication)
  - [Example 3: Composite Authentication](#example-3-composite-authentication)
  - [Example 4: Controllers with Route YAML](#example-4-controllers-with-route-yaml)
  - [Example 5: Full Configuration](#example-5-full-configuration)
- [API Reference](#api-reference)

---

## Overview

`went-web` is a Go library that brings **Symfony-style** security configuration to the Go ecosystem, powered by the **Iris** web framework.

### Key Features

- **Flexible Authentication**: Cookie, Bearer Token, or Composite (chain of methods)
- **Declarative YAML Configuration**: Firewalls, access control, CSRF, sessions
- **Controller System**: Automatic method discovery via reflection
- **User Management**: Interface for RBAC (Role-Based Access Control)
- **CSRF Protection**: Native integration with CSRF middleware for Iris
- **Route Configuration**: Define routes via YAML files

---

## Installation

```bash
go get github.com/zeroSal/went-web
```

---

## Configuration

### security.yaml

The `security.yaml` file defines the security configuration in Symfony style:

```yaml
# config/security.yaml
firewalls:
    main:
        pattern: ^/
        auth:
            cookie:
                name: "session_token"
            bearer:
                enabled: true
        provider: app_user_provider

access:
    - { path: /login, require: IS_AUTHENTICATED_ANONYMOUSLY }
    - { path: /api/*, require: ROLE_API }
    - { path: /*, require: AUTH_REQUIRED }

session:
    cookie: "session_token"
    cookie_path: "/"
    domain: "localhost"
    expires: 3600
    secure: false
    allow_reclaim: true
    disable_subdomain_persistence: false

csrf:
    enabled: true
    secret: "your-csrf-secret-key"
    secure: false
    same_site: "Lax"
    field_name: "_csrf_token"
    header_name: "X-CSRF-Token"

logout:
    enabled: true
    logout_url: /logout
    delete_cookies: ["session_token"]
    redirect_url: /login

entry_point:
    login_url: /login
    code: 401

access_denied:
    enabled: true
    url: /access-denied
```

#### `SecurityConfig` Structure

```go
type SecurityConfig struct {
    Firewalls    map[string]FirewallConfig `yaml:"firewalls"`
    Access       []AccessRule              `yaml:"access"`
    Session      *SessionConfig            `yaml:"session"`
    CSRF         *CSRFConfig               `yaml:"csrf"`
    Logout       *LogoutConfig             `yaml:"logout"`
    EntryPoint   *EntryPointConfig         `yaml:"entry_point"`
    AccessDenied *AccessDeniedConfig       `yaml:"access_denied"`
}
```

### routes.yaml

Defines application routes:

```yaml
# config/routes.yaml
routes:
    - path: /login
      method: ["GET", "POST"]
      handler: AuthController.Login
      require: IS_AUTHENTICATED_ANONYMOUSLY

    - path: /dashboard
      method: ["GET"]
      handler: DashboardController.Index
      require: AUTH_REQUIRED

    - path: /api/users
      method: ["GET", "POST"]
      handlers:
          GET: ApiController.ListUsers
          POST: ApiController.CreateUser
      require: ROLE_API
```

---

## Package `auth` - Authentication

### auth.Interface

Base interface that all authenticators must implement.

```go
package auth

import (
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/session"
    "github.com/zeroSal/went-web/user"
)

type Interface interface {
    // Authenticate attempts to authenticate the current request.
    // Returns the authenticated user and a success boolean.
    Authenticate(ctx iris.Context) (user.Interface, bool)

    // Middleware returns an Iris handler for authentication.
    Middleware() iris.Handler

    // SetSessionProvider sets the provider for loading users.
    SetSessionProvider(provider session.ProviderInterface)
}
```

### Cookie Authentication

Authentication via session cookie.

```go
package auth

type Cookie struct {
    Base
    CookieName      string
    SessionProvider session.ProviderInterface
}

// NewCookie creates a new cookie-based authenticator.
func NewCookie(name string) *Cookie

// SetSessionProvider sets the provider for loading the user.
func (c *Cookie) SetSessionProvider(provider session.ProviderInterface)

// Authenticate verifies the session cookie and loads the user.
func (c *Cookie) Authenticate(ctx iris.Context) (user.Interface, bool)

// Middleware returns the Iris middleware for cookie authentication.
func (b *Cookie) Middleware() iris.Handler
```

**Example:**
```go
cookieAuth := auth.NewCookie("session_token")
cookieAuth.SetSessionProvider(myUserProvider)
app.Use(cookieAuth.Middleware())
```

### Bearer Authentication

Authentication via Bearer Token in the `Authorization` header.

```go
package auth

type Bearer struct {
    Base
    SessionProvider session.ProviderInterface
}

// NewBearer creates a new Bearer authenticator.
func NewBearer() *Bearer

// SetSessionProvider sets the provider for loading the user.
func (b *Bearer) SetSessionProvider(provider session.ProviderInterface)

// Authenticate extracts and verifies the Bearer token.
func (b *Bearer) Authenticate(ctx iris.Context) (user.Interface, bool)

// Middleware returns the Iris middleware for Bearer authentication.
func (b *Bearer) Middleware() iris.Handler
```

**Example:**
```go
bearerAuth := auth.NewBearer()
bearerAuth.SetSessionProvider(myUserProvider)
app.Use(bearerAuth.Middleware())
```

### Composite Authentication

Chains multiple authentication methods: tries each in order until success.

```go
package auth

type Composite struct {
    Base
    authenticators  []Interface
    sessionProvider session.ProviderInterface
}

// NewComposite creates a composite authenticator from a list of authenticators.
func NewComposite(authenticators ...Interface) *Composite

// SetSessionProvider sets the provider on all authenticators in the chain.
func (c *Composite) SetSessionProvider(provider session.ProviderInterface)

// Authenticate tries each authenticator in order.
func (c *Composite) Authenticate(ctx iris.Context) (user.Interface, bool)

// Middleware returns the composite Iris middleware.
func (b *Composite) Middleware() iris.Handler
```

**Example:**
```go
authHandler := auth.NewComposite(
    auth.NewCookie("session_token"),
    auth.NewBearer(),
)
authHandler.SetSessionProvider(myUserProvider)
```

### Base (Utility)

Base struct with shared utility methods for authenticators.

```go
package auth

type Base struct{}

// ExtractToken extracts the Bearer token from the Authorization header.
func (b *Base) ExtractToken(ctx iris.Context) string

// Middleware creates a generic Iris middleware from an authentication function.
func (b *Base) Middleware(
    authenticate func(ctx iris.Context) (user.Interface, bool),
) iris.Handler
```

---

## Package `user` - User Management

Interface for the authenticated user representation.

```go
package user

type Interface interface {
    // GetID returns the user's unique identifier.
    GetID() any

    // GetUsername returns the username.
    GetUsername() string

    // GetRoles returns the assigned roles.
    GetRoles() []string

    // HasRole checks if the user has a specific role.
    HasRole(role string) bool
}
```

**Implementation Example:**
```go
type User struct {
    ID       int
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
    for _, r := range u.Roles {
        if r == role {
            return true
        }
    }
    return false
}
```

---

## Package `session` - Session Provider

Interface for loading a user from credentials/tokens.

```go
package session

import "github.com/zeroSal/went-web/user"

type ProviderInterface interface {
    // Load loads a user from the provided credentials.
    // credential can be a token, session ID, etc.
    Load(credential any) (user.Interface, error)
}
```

**Implementation Example:**
```go
type UserProvider struct {
    db *sql.DB
}

func (p *UserProvider) Load(credential any) (user.Interface, error) {
    token, ok := credential.(string)
    if !ok {
        return nil, errors.New("invalid credential type")
    }

    // Load user from database using the token
    var u User
    err := p.db.QueryRow("SELECT id, username FROM users WHERE token = ?", token).
        Scan(&u.ID, &u.Username)
    if err != nil {
        return nil, err
    }

    return &u, nil
}
```

---

## Package `controller` - Controllers

Controller system with automatic method discovery via reflection.

### Interfaces and Structs

```go
package controller

// Base is the base struct for controllers.
type Base struct{}

// Configuration defines the controller name.
type Configuration struct {
    Name string
}

// Interface that every controller must implement.
type Interface interface {
    // GetConfiguration returns the controller configuration.
    GetConfiguration() Configuration
}
```

### Registry

The Registry automatically discovers controller handler methods via reflection.

```go
package controller

type Handler struct {
    Controller string
    Method     string
    Handler    iris.Handler
}

type Registry struct {
    controllers []Interface
}

// NewRegistry creates a new controller registry.
func NewRegistry() *Registry

// Register registers a controller in the registry.
func (r *Registry) Register(c Interface)

// Handlers returns the list of all discovered handlers.
// Methods are automatically discovered if:
// - They accept exactly one parameter of type iris.Context
// - They are not the GetConfiguration method
func (r *Registry) Handlers() []Handler
```

**Controller Example:**
```go
package controllers

import (
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/controller"
)

type AuthController struct {
    controller.Base
}

func (c *AuthController) GetConfiguration() controller.Configuration {
    return controller.Configuration{
        Name: "AuthController",
    }
}

func (c *AuthController) Login(ctx iris.Context) {
    ctx.JSON(iris.Map{"message": "Login page"})
}

func (c *AuthController) Logout(ctx iris.Context) {
    ctx.JSON(iris.Map{"message": "Logged out"})
}
```

**Registering Controllers:**
```go
reg := controller.NewRegistry()
reg.Register(&controllers.AuthController{})
reg.Register(&controllers.DashboardController{})

handlers := reg.Handlers()
// handlers contains: [
//   {Controller: "AuthController", Method: "Login", Handler: func},
//   {Controller: "AuthController", Method: "Logout", Handler: func},
//   ...
// ]
```

---

## Package `security` - Core Security

The main package that orchestrates all security components.

### Security Struct

```go
package security

type Security struct {
    config          *SecurityConfig
    routes          []RouteConfig
    authenticator   auth.Interface
    publicPatterns  []string
    handlerRegistry HandlerRegistry
    session         *sessions.Sessions
    csrf            iris.Handler
    sessionProvider session.ProviderInterface
}
```

### Constructors

```go
// NewSecurity creates a Security instance by reading configuration from a file.
func NewSecurity(configPath string) (*Security, error)

// NewSecurityFromEmbed creates a Security instance from an embedded filesystem.
func NewSecurityFromEmbed(efs embed.FS, path string) (*Security, error)

// NewSecurityFromConfig creates a Security instance from an existing configuration.
func NewSecurityFromConfig(config *SecurityConfig) (*Security, error)
```

### Main Methods

```go
// SetSessionProvider sets the provider for loading users.
func (s *Security) SetSessionProvider(provider session.ProviderInterface)

// GetUserProvider returns the currently configured provider.
func (s *Security) GetUserProvider() session.ProviderInterface

// Middleware returns the main security middleware.
// It handles access control checks based on security.yaml.
func (s *Security) Middleware() iris.Handler

// Authenticator returns the configured authenticator.
func (s *Security) Authenticator() auth.Interface

// GetSessionManager returns the Iris session manager.
func (s *Security) GetSessionManager() *sessions.Sessions

// GetCSRFMiddleware returns the CSRF middleware if enabled.
func (s *Security) GetCSRFMiddleware() iris.Handler

// SetRoutes sets the routes from YAML configuration.
func (s *Security) SetRoutes(routes []RouteConfig)

// RegisterHandler manually registers a handler.
func (s *Security) RegisterHandler(controller, method string, handler iris.Handler)

// RegisterRoutes registers all routes on the Iris app.
// Automatically handles HTTP methods and security middleware.
func (s *Security) RegisterRoutes(app *iris.Application)
```

### HandlerRegistry

```go
package security

type HandlerRegistry map[string]map[string]iris.Handler

func NewHandlerRegistry() HandlerRegistry

func (r HandlerRegistry) Register(controller string, method string, handler iris.Handler)

func (r HandlerRegistry) Get(controller, method string) iris.Handler

func (r HandlerRegistry) Range(fn func(controller, method string, handler iris.Handler))
```

### YAML Configuration - Data Structures

#### FirewallConfig
```go
type FirewallConfig struct {
    Pattern  string     `yaml:"pattern"`
    Auth     AuthConfig `yaml:"auth"`
    Provider string     `yaml:"provider"`
}

type AuthConfig struct {
    Cookie *CookieAuthConfig `yaml:"cookie"`
    Bearer *BearerAuthConfig `yaml:"bearer"`
    JWT    *JWTAuthConfig    `yaml:"jwt"`
}

type CookieAuthConfig struct {
    Name string `yaml:"name"`
}

type BearerAuthConfig struct {
    Enabled bool `yaml:"enabled"`
}

type JWTAuthConfig struct {
    Secret string   `yaml:"secret"`
    Expiry Duration `yaml:"expiry"`
}
```

#### AccessRule
```go
type AccessRule struct {
    Path    string `yaml:"path"`
    Require string `yaml:"require"`
}
```

Valid values for `require`:
- `IS_AUTHENTICATED_ANONYMOUSLY` - Public access
- `AUTH_REQUIRED` - Requires authentication
- `ROLE_*` - Requires a specific role (e.g., `ROLE_ADMIN`)

#### SessionConfig
```go
type SessionConfig struct {
    Cookie                      string `yaml:"cookie"`
    CookiePath                  string `yaml:"cookie_path"`
    Domain                      string `yaml:"domain"`
    Expires                     int    `yaml:"expires"`      // seconds
    Secure                      bool   `yaml:"secure"`
    AllowReclaim                bool   `yaml:"allow_reclaim"`
    DisableSubdomainPersistence bool   `yaml:"disable_subdomain_persistence"`
}
```

#### CSRFConfig
```go
type CSRFConfig struct {
    Enabled    bool   `yaml:"enabled"`
    Secret     string `yaml:"secret"`
    Secure     bool   `yaml:"secure"`
    SameSite   string `yaml:"same_site"`    // "Lax", "Strict", "None"
    FieldName  string `yaml:"field_name"`   // default: "_csrf_token"
    HeaderName string `yaml:"header_name"`  // default: "X-CSRF-Token"
}
```

#### RouteConfig
```go
type RouteConfig struct {
    Path     string            `yaml:"path"`
    Method   []string          `yaml:"method"`    // ["GET", "POST", ...]
    Handler  string            `yaml:"handler"`   // "Controller.Method"
    Handlers map[string]string `yaml:"handlers"`  // {"GET": "Controller.Method1", "POST": "Controller.Method2"}
    Require  string            `yaml:"require"`
}
```

#### Duration
```go
// Duration handles duration in seconds from YAML parsing.
type Duration struct {
    Duration int `yaml:"duration"`
}

func (d Duration) ToDuration() time.Duration
```

#### LogoutConfig
```go
type LogoutConfig struct {
    Enabled       bool     `yaml:"enabled"`
    LogoutUrl     string   `yaml:"logout_url"`
    DeleteCookies []string `yaml:"delete_cookies"`
    RedirectUrl   string   `yaml:"redirect_url"`
}
```

#### EntryPointConfig
```go
type EntryPointConfig struct {
    LoginUrl string `yaml:"login_url"`
    Code     int    `yaml:"code"`
}
```

#### AccessDeniedConfig
```go
type AccessDeniedConfig struct {
    Enabled bool   `yaml:"enabled"`
    Url     string `yaml:"url"`
}
```

### Configuration Loading Functions

```go
// LoadSecurityConfig loads security configuration from a file.
func LoadSecurityConfig(path string) (*SecurityConfig, error)

// LoadSecurityConfigFromBytes loads from YAML bytes.
func LoadSecurityConfigFromBytes(data []byte) (*SecurityConfig, error)

// LoadSecurityConfigFromEmbed loads from an embedded filesystem.
func LoadSecurityConfigFromEmbed(efs embed.FS, path string) (*SecurityConfig, error)

// LoadRoutesConfig loads route configuration from a file.
func LoadRoutesConfig(path string) ([]RouteConfig, error)

// LoadRoutesConfigFromBytes loads from YAML bytes.
func LoadRoutesConfigFromBytes(data []byte) ([]RouteConfig, error)

// LoadRoutesConfigFromEmbed loads from an embedded filesystem.
func LoadRoutesConfigFromEmbed(efs embed.FS, path string) ([]RouteConfig, error)
```

---

## Complete Examples

### Example 1: Cookie Authentication

**config/security.yaml:**
```yaml
firewalls:
    main:
        pattern: ^/
        auth:
            cookie:
                name: "session_token"

access:
    - { path: /login, require: IS_AUTHENTICATED_ANONYMOUSLY }
    - { path: /*, require: AUTH_REQUIRED }
```

**main.go:**
```go
package main

import (
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/security"
)

type MyUserProvider struct{}

func (p *MyUserProvider) Load(credential any) (user.Interface, error) {
    // Implement user loading from token
    return nil, nil
}

func main() {
    app := iris.New()
    sec, _ := security.NewSecurity("config/security.yaml")

    sec.SetSessionProvider(&MyUserProvider{})

    app.Use(sec.Middleware())

    app.Get("/login", func(ctx iris.Context) {
        ctx.HTML("<h1>Login Page</h1>")
    })

    app.Get("/dashboard", func(ctx iris.Context) {
        ctx.HTML("<h1>Dashboard</h1>")
    })

    app.Listen(":8080")
}
```

### Example 2: Bearer Token Authentication

**config/security.yaml:**
```yaml
firewalls:
    api:
        pattern: ^/api
        auth:
            bearer:
                enabled: true

access:
    - { path: /api/*, require: ROLE_API }
```

**main.go:**
```go
package main

import (
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/security"
)

func main() {
    app := iris.New()
    sec, _ := security.NewSecurity("config/security.yaml")

    sec.SetSessionProvider(&MyUserProvider{})

    app.Use(sec.Middleware())

    api := app.Party("/api")
    {
        api.Get("/users", func(ctx iris.Context) {
            ctx.JSON(iris.Map{"users": []string{}})
        })
    }

    app.Listen(":8080")
}
```

### Example 3: Composite Authentication

**config/security.yaml:**
```yaml
firewalls:
    main:
        pattern: ^/
        auth:
            cookie:
                name: "session_token"
            bearer:
                enabled: true

access:
    - { path: /public/*, require: IS_AUTHENTICATED_ANONYMOUSLY }
    - { path: /api/*, require: ROLE_API }
    - { path: /*, require: AUTH_REQUIRED }
```

**main.go:**
```go
package main

import (
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/auth"
    "github.com/zeroSal/went-web/security"
)

func main() {
    app := iris.New()

    // Create composite authenticator
    authenticator := auth.NewComposite(
        auth.NewCookie("session_token"),
        auth.NewBearer(),
    )
    authenticator.SetSessionProvider(&MyUserProvider{})

    sec, _ := security.NewSecurity("config/security.yaml")
    sec.SetSessionProvider(&MyUserProvider{})

    app.Use(sec.Middleware())

    app.Listen(":8080")
}
```

### Example 4: Controllers with Route YAML

**config/routes.yaml:**
```yaml
routes:
    - path: /login
      method: ["GET", "POST"]
      handler: AuthController.Login
      require: IS_AUTHENTICATED_ANONYMOUSLY

    - path: /dashboard
      method: ["GET"]
      handler: DashboardController.Index
      require: AUTH_REQUIRED
```

**controllers/auth_controller.go:**
```go
package controllers

import (
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/controller"
)

type AuthController struct {
    controller.Base
}

func (c *AuthController) GetConfiguration() controller.Configuration {
    return controller.Configuration{Name: "AuthController"}
}

func (c *AuthController) Login(ctx iris.Context) {
    ctx.JSON(iris.Map{"message": "Login"})
}
```

**controllers/dashboard_controller.go:**
```go
package controllers

import (
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/controller"
)

type DashboardController struct {
    controller.Base
}

func (c *DashboardController) GetConfiguration() controller.Configuration {
    return controller.Configuration{Name: "DashboardController"}
}

func (c *DashboardController) Index(ctx iris.Context) {
    ctx.JSON(iris.Map{"message": "Dashboard"})
}
```

**main.go:**
```go
package main

import (
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/controller"
    "github.com/zeroSal/went-web/security"
)

func main() {
    app := iris.New()

    // Load security configuration
    sec, _ := security.NewSecurity("config/security.yaml")
    sec.SetSessionProvider(&MyUserProvider{})

    // Register controllers
    reg := controller.NewRegistry()
    reg.Register(&controllers.AuthController{})
    reg.Register(&controllers.DashboardController{})

    // Register handlers in security
    for _, h := range reg.Handlers() {
        sec.RegisterHandler(h.Controller, h.Method, h.Handler)
    }

    // Load routes from YAML
    routes, _ := security.LoadRoutesConfig("config/routes.yaml")
    sec.SetRoutes(routes)
    sec.RegisterRoutes(app)

    app.Listen(":8080")
}
```

### Example 5: Full Configuration with CSRF and Sessions

**config/security.yaml:**
```yaml
firewalls:
    main:
        pattern: ^/
        auth:
            cookie:
                name: "session_token"
        provider: app_user_provider

access:
    - { path: /login, require: IS_AUTHENTICATED_ANONYMOUSLY }
    - { path: /register, require: IS_AUTHENTICATED_ANONYMOUSLY }
    - { path: /admin/*, require: ROLE_ADMIN }
    - { path: /*, require: AUTH_REQUIRED }

session:
    cookie: "session_token"
    cookie_path: "/"
    domain: "localhost"
    expires: 3600
    secure: false
    allow_reclaim: true

csrf:
    enabled: true
    secret: "my-secret-csrf-key"
    secure: false
    same_site: "Lax"
    field_name: "_csrf_token"
    header_name: "X-CSRF-Token"

logout:
    enabled: true
    logout_url: /logout
    delete_cookies: ["session_token"]
    redirect_url: /login
```

**main.go:**
```go
package main

import (
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/security"
)

func main() {
    app := iris.New()

    sec, err := security.NewSecurity("config/security.yaml")
    if err != nil {
        panic(err)
    }

    sec.SetSessionProvider(&MyUserProvider{})

    // Security middleware
    app.Use(sec.Middleware())

    // CSRF middleware (if enabled)
    if csrf := sec.GetCSRFMiddleware(); csrf != nil {
        app.Use(csrf)
    }

    // Public routes
    app.Get("/login", func(ctx iris.Context) {
        ctx.HTML("<h1>Login</h1>")
    })

    // Protected routes
    app.Get("/dashboard", func(ctx iris.Context) {
        ctx.HTML("<h1>Dashboard</h1>")
    })

    // Admin routes
    app.Get("/admin", func(ctx iris.Context) {
        ctx.HTML("<h1>Admin Panel</h1>")
    })

    // Logout
    app.Get("/logout", func(ctx iris.Context) {
        sessionManager := sec.GetSessionManager()
        if sessionManager != nil {
            sessionManager.Destroy(ctx)
        }
        ctx.Redirect("/login")
    })

    app.Listen(":8080")
}
```

---

## API Reference

### Package `auth`

| Type | Name | Description |
|------|------|-------------|
| Interface | `Interface` | Base interface for authenticators |
| Struct | `Base` | Utility for authenticators |
| Struct | `Cookie` | Cookie-based authentication |
| Struct | `Bearer` | Bearer token authentication |
| Struct | `Composite` | Composite (chain) authentication |
| Func | `NewCookie(name)` | Creates cookie authenticator |
| Func | `NewBearer()` | Creates bearer authenticator |
| Func | `NewComposite(...)` | Creates composite authenticator |

### Package `user`

| Type | Name | Description |
|------|------|-------------|
| Interface | `Interface` | Contract for user objects |
| Method | `GetID()` | Returns user ID |
| Method | `GetUsername()` | Returns username |
| Method | `GetRoles()` | Returns roles |
| Method | `HasRole(role)` | Checks role membership |

### Package `session`

| Type | Name | Description |
|------|------|-------------|
| Interface | `ProviderInterface` | Provider for loading user from credentials |
| Method | `Load(credential)` | Loads user from token/credentials |

### Package `controller`

| Type | Name | Description |
|------|------|-------------|
| Struct | `Base` | Base controller struct |
| Struct | `Configuration` | Controller name configuration |
| Interface | `Interface` | Controller interface |
| Struct | `Handler` | Controller/method/handler mapping |
| Struct | `Registry` | Registry for controller discovery |
| Func | `NewRegistry()` | Creates new registry |
| Method | `Register(c)` | Registers a controller |
| Method | `Handlers()` | Returns discovered handlers |

### Package `security`

| Type | Name | Description |
|------|------|-------------|
| Struct | `Security` | Main security component |
| Struct | `SecurityConfig` | Complete configuration |
| Struct | `FirewallConfig` | Firewall configuration |
| Struct | `AccessRule` | Access rule |
| Struct | `SessionConfig` | Session configuration |
| Struct | `CSRFConfig` | CSRF configuration |
| Struct | `RouteConfig` | Route configuration |
| Struct | `HandlerRegistry` | Handler registry |
| Func | `NewSecurity(path)` | Creates from file |
| Func | `NewSecurityFromEmbed()` | Creates from embed.FS |
| Func | `NewSecurityFromConfig()` | Creates from existing config |
| Method | `Middleware()` | Main middleware |
| Method | `RegisterRoutes(app)` | Registers routes on Iris |
| Method | `SetSessionProvider()` | Sets user provider |
| Func | `LoadSecurityConfig()` | Loads security.yaml |
| Func | `LoadRoutesConfig()` | Loads routes.yaml |

---

## Makefile Commands

```bash
make test              # Run unit tests with coverage
make test-coverage     # Show coverage summary
make coverage-html     # Generate HTML coverage report
make lint              # Run linter
make examples          # Compile all examples
make test-all          # Unit tests + functional tests on examples
make run-01-cookie-auth  # Run specific example
make clean             # Clean compiled files
```
