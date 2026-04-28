# Authentication System

## Interface

All authenticators implement `auth.Interface`:

```go
type Interface interface {
    Authenticate(ctx iris.Context) (interface{}, bool) // Returns user & success flag
    Middleware() iris.Handler                      // Returns Iris middleware
}
```

## Cookie Authentication

Authenticates users based on a cookie value.

```go
import "github.com/zeroSal/went-web/auth"

// Create authenticator - checks for "SESSION_ID" cookie
cookieAuth := auth.NewCookie("SESSION_ID")

// Use in handler
if user, ok := cookieAuth.Authenticate(ctx); ok {
    // user is the cookie value (string)
    fmt.Println("Authenticated:", user)
}

// Or use as middleware
app.Get("/profile", cookieAuth.Middleware(), func(ctx iris.Context) {
    ctx.WriteString("Profile page")
})
```

**YAML config:**
```yaml
firewalls:
  main:
    pattern: "^/"
    auth:
      cookie:
        name: "SESSION_ID"    # Cookie name to check
```

## Bearer Authentication

Extracts and validates Bearer tokens from `Authorization` header.

```go
import "github.com/zeroSal/went-web/auth"

// Create authenticator
bearerAuth := auth.NewBearer()

// Use in handler
if token, ok := bearerAuth.Authenticate(ctx); ok {
    // token is the bearer token string
    fmt.Println("Token:", token)
}
```

**YAML config:**
```yaml
firewalls:
  main:
    pattern: "^/"
    auth:
      bearer:
        enabled: true
```

## JWT Authentication

Handles JWT (JSON Web Token) with validation and token generation.

```go
import (
    "github.com/zeroSal/went-web/auth"
    "time"
)

// Create authenticator with secret and expiry
jwtAuth := auth.NewJWT(
    []byte("your-32-byte-secret-key-here!!!!"),
    time.Hour, // Token expiry
)

// Generate token (e.g., in login handler)
claims := jwt.MapClaims{
    "sub":      "user123",      // User ID
    "username": "john",         // Username
    "roles":    []string{"admin", "user"}, // Roles
}
token, err := jwtAuth.GenerateToken(claims)
if err != nil {
    panic(err)
}

// Authenticate in middleware
if claims, ok := jwtAuth.Authenticate(ctx); ok {
    // claims is jwt.MapClaims
    fmt.Println("User ID:", claims["sub"])
    fmt.Println("Username:", claims["username"])
}
```

**YAML config:**
```yaml
firewalls:
  main:
    pattern: "^/"
    auth:
      jwt:
        secret: "your-32-byte-secret-key-here!!!!"  # Must be 32+ bytes
        expiry: 3600                                 # Seconds
```

**JWT Flow:**
1. Client sends `Authorization: Bearer <token>` header
2. Server validates token signature with secret
3. Checks token expiry
4. Returns `jwt.MapClaims` on success

## Composite Authentication

Combines multiple strategies. Tries each in order, returns on first success.

```go
import "github.com/zeroSal/went-web/auth"

// Create composite - tries Cookie, then Bearer, then JWT
compositeAuth := auth.NewComposite(
    auth.NewCookie("SESSION_ID"),
    auth.NewBearer(),
    auth.NewJWT([]byte("secret"), time.Hour),
)

// Use as middleware
app.Use(compositeAuth.Middleware())
```

**YAML config (auto-creates Composite):**
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
        secret: "your-secret"
        expiry: 3600
```

## Using with Security

```go
import (
    "github.com/zeroSal/went-web/security"
    "github.com/zeroSal/went-web/auth"
)

sec, _ := security.NewSecurity("security.yaml")

// Get authenticator (type assert for specific methods)
authenticator := sec.Authenticator()

// Access JWT-specific methods
if jwtAuth, ok := authenticator.(*auth.JWT); ok {
    token, _ := jwtAuth.GenerateToken(claims)
}
```

## Custom Authenticators

Implement `auth.Interface` for custom auth:

```go
import (
    "github.com/kataras/iris/v12"
    "github.com/zeroSal/went-web/auth"
)

type APIKeyAuth struct {
    HeaderName string
    ValidKeys   map[string]string // key -> userID
}

func (a *APIKeyAuth) Authenticate(ctx iris.Context) (interface{}, bool) {
    apiKey := ctx.GetHeader(a.HeaderName)
    if apiKey == "" {
        return nil, false
    }

    if userID, ok := a.ValidKeys[apiKey]; ok {
        return userID, true
    }

    return nil, false
}

func (a *APIKeyAuth) Middleware() iris.Handler {
    return func(ctx iris.Context) {
        if _, ok := a.Authenticate(ctx); !ok {
            ctx.StatusCode(iris.StatusUnauthorized)
            ctx.StopExecution()
            return
        }
        ctx.Next()
    }
}
```
