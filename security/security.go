package security

import (
	"embed"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	csrf "github.com/iris-contrib/middleware/csrf"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"github.com/zeroSal/went-web/auth"
	"github.com/zeroSal/went-web/session"
)

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

func NewSecurity(configPath string) (*Security, error) {
	config, err := LoadSecurityConfig(configPath)
	if err != nil {
		return nil, err
	}
	return newSecurityFromConfig(config)
}

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

func NewSecurityFromConfig(config *SecurityConfig) (*Security, error) {
	return newSecurityFromConfig(config)
}

func newSecurityFromConfig(config *SecurityConfig) (*Security, error) {
	auths := []auth.Interface{}

	for _, fw := range config.Firewalls {
		if fw.Auth.Cookie != nil {
			auths = append(auths, auth.NewCookie(fw.Auth.Cookie.Name))
		}

		if fw.Auth.Bearer != nil && fw.Auth.Bearer.Enabled {
			auths = append(auths, auth.NewBearer())
		}

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
			return nil, errors.New("CSRF secret is required")
		}

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

func (s *Security) SetSessionProvider(provider session.ProviderInterface) {
	s.sessionProvider = provider
	if s.authenticator != nil {
		if p, ok := s.authenticator.(interface {
			SetUserProvider(session.ProviderInterface)
		}); ok {
			p.SetUserProvider(provider)
		}
	}
}

func (s *Security) GetUserProvider() session.ProviderInterface {
	return s.sessionProvider
}

func (s *Security) Middleware() iris.Handler {
	return func(ctx iris.Context) {
		path := ctx.Path()
		method := ctx.Method()

		for _, rule := range s.config.Access {
			if !matchesPath(rule.Path, path, method) {
				continue
			}

			if s.handleAccessCheck(ctx, rule.Require) {
				ctx.Next()
			}
			return
		}

		isPublic := false
		for _, pattern := range s.publicPatterns {
			if matchesPath(pattern, path, method) {
				isPublic = true
				break
			}
		}

		if !isPublic {
			if s.handleAccessCheck(ctx, "AUTH_REQUIRED") {
				ctx.Next()
			}
			return
		}

		ctx.Next()
	}
}

func (s *Security) Authenticator() auth.Interface {
	return s.authenticator
}

func (s *Security) GetSessionManager() *sessions.Sessions {
	return s.session
}

func (s *Security) GetCSRFMiddleware() iris.Handler {
	return s.csrf
}

func (s *Security) SetRoutes(routes []RouteConfig) {
	s.routes = routes

	for _, route := range routes {
		if route.Require == "IS_AUTHENTICATED_ANONYMOUSLY" {
			s.publicPatterns = append(s.publicPatterns, route.Path)
		}
	}
}

func (s *Security) RegisterHandler(controller, method string, handler iris.Handler) {
	s.handlerRegistry.Register(controller, method, handler)
}

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

func (s *Security) handleAccessCheck(ctx iris.Context, require string) bool {
	if require == "IS_AUTHENTICATED_ANONYMOUSLY" {
		return true
	}

	if s.authenticator == nil {
		return true
	}

	user, ok := s.authenticator.Authenticate(ctx)
	if !ok {
		if s.config.EntryPoint != nil && s.config.EntryPoint.LoginUrl != "" {
			ctx.Redirect(s.config.EntryPoint.LoginUrl, s.config.EntryPoint.Code)
			return false
		}

		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.StopExecution()
		return false
	}

	if require == "AUTH_REQUIRED" {
		return true
	}

	if strings.HasPrefix(require, "ROLE_") {
		if !user.HasRole(require) {
			ctx.StatusCode(iris.StatusForbidden)
			ctx.StopExecution()
			return false
		}
		return true
	}

	return true
}

func (s *Security) buildMiddleware(require string) iris.Handler {
	return func(ctx iris.Context) {
		if s.handleAccessCheck(ctx, require) {
			ctx.Next()
		}
	}
}

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
