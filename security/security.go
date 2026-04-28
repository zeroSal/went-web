package security

import (
	"embed"
	"net/http"
	"regexp"
	"strings"
	"time"

	csrf "github.com/iris-contrib/middleware/csrf"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"github.com/zeroSal/went-web/auth"
	"github.com/zeroSal/went-web/user"
)

// HandlerRegistry maps controller/method combinations to Iris handlers
type HandlerRegistry map[string]map[string]iris.Handler

// NewHandlerRegistry creates a new empty HandlerRegistry
func NewHandlerRegistry() HandlerRegistry {
	return make(HandlerRegistry)
}

// Register adds a handler for a controller and method
func (r HandlerRegistry) Register(controller string, method string, handler iris.Handler) {
	if r[controller] == nil {
		r[controller] = make(map[string]iris.Handler)
	}
	r[controller][method] = handler
}

// Get retrieves a handler for a controller and method
func (r HandlerRegistry) Get(controller, method string) iris.Handler {
	if r[controller] != nil {
		return r[controller][method]
	}
	return nil
}

// Range iterates over all registered handlers
func (r HandlerRegistry) Range(fn func(controller, method string, handler iris.Handler)) {
	for controller, methods := range r {
		for method, handler := range methods {
			fn(controller, method, handler)
		}
	}
}

// Security manages authentication, authorization, and route security
type Security struct {
	config          *SecurityConfig
	routes          []RouteConfig
	authenticator   auth.Interface
	publicPatterns  []string
	handlerRegistry HandlerRegistry
	session         *sessions.Sessions
	csrf            iris.Handler
	userProvider    user.Provider
	roleChecker     user.RoleChecker
}

// NewSecurity creates a new Security instance from a config file path
func NewSecurity(configPath string) (*Security, error) {
	config, err := LoadSecurityConfig(configPath)
	if err != nil {
		return nil, err
	}
	return newSecurityFromConfig(config)
}

// NewSecurityFromEmbed creates a new Security instance from an embedded filesystem
func NewSecurityFromEmbed(efs embed.FS, path string) (*Security, error) {
	data, err := efs.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config, err := LoadSecurityConfigFromBytes(data)
	if err != nil {
		return nil, err
	}
	return newSecurityFromConfig(config)
}

// NewSecurityFromConfig creates a new Security instance from a config struct
func NewSecurityFromConfig(config *SecurityConfig) (*Security, error) {
	return newSecurityFromConfig(config)
}

// newSecurityFromConfig builds a Security instance from config
func newSecurityFromConfig(config *SecurityConfig) (*Security, error) {
	auths := []auth.Interface{}

	for _, fw := range config.Firewalls {
		if fw.Auth.Cookie != nil {
			auths = append(auths, auth.NewCookie(fw.Auth.Cookie.Name))
		}
		if fw.Auth.Bearer != nil && fw.Auth.Bearer.Enabled {
			auths = append(auths, auth.NewBearer())
		}
		if fw.Auth.JWT != nil {
			auths = append(auths, auth.NewJWT([]byte(fw.Auth.JWT.Secret), time.Duration(fw.Auth.JWT.Expiry.Duration)*time.Second))
		}

		// If session is configured, add cookie auth for session cookie
		if config.Session != nil && config.Session.Cookie != "" {
			auths = append(auths, auth.NewCookie(config.Session.Cookie))
		}
	}

	var authenticator auth.Interface
	if len(auths) == 1 {
		authenticator = auths[0]
	} else if len(auths) > 1 {
		authenticator = auth.NewComposite(auths...)
	}

	publicPatterns := []string{}
	for _, rule := range config.Access {
		if rule.Require == "IS_AUTHENTICATED_ANONYMOUSLY" {
			publicPatterns = append(publicPatterns, rule.Path)
		}
	}

	sec := &Security{
		config:          config,
		authenticator:   authenticator,
		publicPatterns:  publicPatterns,
		handlerRegistry: NewHandlerRegistry(),
	}

	if config.Session != nil {
		sessConfig := sessions.Config{
			Cookie:                      config.Session.Cookie,
			CookieSecureTLS:             config.Session.Secure,
			AllowReclaim:                config.Session.AllowReclaim,
			DisableSubdomainPersistence: config.Session.DisableSubdomainPersistence,
		}
		if config.Session.Expires > 0 {
			sessConfig.Expires = time.Duration(config.Session.Expires) * time.Second
		}
		sec.session = sessions.New(sessConfig)
	}

	if config.CSRF != nil && config.CSRF.Enabled {
		csrfKey := config.CSRF.Secret
		if csrfKey == "" {
			csrfKey = "32-byte-long-auth-key-here!!!!"
		}

		// Build cookie options
		var cookieOpts []csrf.CookieOption
		if config.CSRF.Secure {
			cookieOpts = append(cookieOpts, csrf.Secure(true))
		}
		if config.CSRF.SameSite != "" {
			switch config.CSRF.SameSite {
			case "Lax":
				cookieOpts = append(cookieOpts, csrf.SameSite(http.SameSiteLaxMode))
			case "Strict":
				cookieOpts = append(cookieOpts, csrf.SameSite(http.SameSiteStrictMode))
			case "None":
				cookieOpts = append(cookieOpts, csrf.SameSite(http.SameSiteNoneMode))
			}
		}

		// Build options for New() to support FieldName and HeaderName
		opts := csrf.Options{
			FieldName:     config.CSRF.FieldName,
			RequestHeader: config.CSRF.HeaderName,
		}
		if len(cookieOpts) > 0 {
			opts.Store = csrf.NewCookieStore([]byte(csrfKey), cookieOpts...)
		} else {
			opts.Store = csrf.NewCookieStore([]byte(csrfKey))
		}

		csrfInstance := csrf.New(opts)
		sec.csrf = iris.Handler(csrfInstance.Protect)
	}

	return sec, nil
}

// SetUserProvider sets the user provider for loading users
func (s *Security) SetUserProvider(provider user.Provider) {
	s.userProvider = provider
}

// SetRoleChecker sets the role checker for authorization
func (s *Security) SetRoleChecker(checker user.RoleChecker) {
	s.roleChecker = checker
}

// UserProvider returns the current user provider
func (s *Security) UserProvider() user.Provider {
	return s.userProvider
}

// RoleChecker returns the current role checker
func (s *Security) RoleChecker() user.RoleChecker {
	return s.roleChecker
}

// Middleware returns the main security middleware
func (s *Security) Middleware() iris.Handler {
	return func(ctx iris.Context) {
		path := ctx.Path()
		method := ctx.Method()

		isPublic := false
		for _, pattern := range s.publicPatterns {
			if matchesPath(pattern, path, method) {
				isPublic = true
				break
			}
		}

		if !isPublic && s.authenticator != nil {
			if _, ok := s.authenticator.Authenticate(ctx); !ok {
				ctx.StatusCode(iris.StatusUnauthorized)
				ctx.StopExecution()
				return
			}
		}

		ctx.Next()
	}
}

// Authenticator returns the current authenticator
func (s *Security) Authenticator() auth.Interface {
	return s.authenticator
}

// Session returns the session manager
func (s *Security) Session() *sessions.Sessions {
	return s.session
}

// CSRF returns the CSRF middleware
func (s *Security) CSRF() iris.Handler {
	return s.csrf
}

// SetRoutes sets the route configurations
func (s *Security) SetRoutes(routes []RouteConfig) {
	s.routes = routes

	for _, route := range routes {
		if route.Require == "IS_AUTHENTICATED_ANONYMOUSLY" {
			s.publicPatterns = append(s.publicPatterns, route.Path)
		}
	}
}

// RegisterHandler registers a handler for a controller and method
func (s *Security) RegisterHandler(controller, method string, handler iris.Handler) {
	s.handlerRegistry.Register(controller, method, handler)
}

// RegisterRoutes registers all configured routes to the Iris application
func (s *Security) RegisterRoutes(app *iris.Application) {
	for _, route := range s.routes {
		middleware := s.buildMiddleware(route.Require)

		methods := route.Method
		if len(methods) == 0 {
			methods = []string{"GET"}
		}

		if route.Handlers != nil {
			for _, method := range methods {
				methodUpper := strings.ToUpper(method)
				if handlerName, ok := route.Handlers[methodUpper]; ok {
					handler := s.resolveHandler(handlerName)
					if handler == nil {
						continue
					}
					switch methodUpper {
					case "GET":
						app.Get(route.Path, middleware, handler)
					case "POST":
						app.Post(route.Path, middleware, handler)
					case "PUT":
						app.Put(route.Path, middleware, handler)
					case "DELETE":
						app.Delete(route.Path, middleware, handler)
					case "PATCH":
						app.Patch(route.Path, middleware, handler)
					}
				}
			}
			continue
		}

		handler := s.resolveHandler(route.Handler)
		if handler == nil {
			continue
		}

		for _, method := range methods {
			switch strings.ToUpper(method) {
			case "GET":
				app.Get(route.Path, middleware, handler)
			case "POST":
				app.Post(route.Path, middleware, handler)
			case "PUT":
				app.Put(route.Path, middleware, handler)
			case "DELETE":
				app.Delete(route.Path, middleware, handler)
			case "PATCH":
				app.Patch(route.Path, middleware, handler)
			}
		}
	}
}

// resolveHandler looks up a handler by "Controller.Method" string
func (s *Security) resolveHandler(handler string) iris.Handler {
	parts := strings.Split(handler, ".")
	if len(parts) != 2 {
		return nil
	}
	controller := parts[0]
	method := parts[1]
	methodLower := strings.ToLower(method)

	for ctrl, methods := range s.handlerRegistry {
		if !strings.EqualFold(ctrl, controller) {
			continue
		}
		for m, h := range methods {
			if strings.ToLower(m) == methodLower {
				return h
			}
		}
	}
	return nil
}

// buildMiddleware creates middleware for a specific require rule
func (s *Security) buildMiddleware(require string) iris.Handler {
	return func(ctx iris.Context) {
		isPublic := require == "IS_AUTHENTICATED_ANONYMOUSLY"

		if !isPublic && s.authenticator != nil {
			if _, ok := s.authenticator.Authenticate(ctx); !ok {
				ctx.StatusCode(iris.StatusUnauthorized)
				ctx.StopExecution()
				return
			}
		}

		ctx.Next()
	}
}

// matchesPath checks if a path matches a pattern with optional method prefix
func matchesPath(pattern, path, method string) bool {
	if strings.Contains(pattern, " ") {
		parts := strings.SplitN(pattern, " ", 2)
		patternMethod := strings.TrimSpace(parts[0])
		pattern = strings.TrimSpace(parts[1])
		if !strings.EqualFold(patternMethod, method) {
			return false
		}
	}

	if before, ok := strings.CutSuffix(pattern, "*"); ok {
		if strings.HasPrefix(path, before) {
			return true
		}
	}

	pathPattern := pattern
	if strings.Contains(pattern, "{") {
		pathPattern = strings.ReplaceAll(pattern, "{id}", `[^/]+`)
		pathPattern = strings.ReplaceAll(pathPattern, "{slug}", `[^/]+`)
	}

	matched, err := regexp.MatchString("^"+pathPattern+"$", path)
	return err == nil && matched
}
