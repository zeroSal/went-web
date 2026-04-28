package security

import (
	"testing"

	"github.com/kataras/iris/v12"
	"github.com/zeroSal/went-web/user"
)

func TestLoadSecurityConfig(t *testing.T) {
	yaml := []byte(`
firewalls:
  main:
    pattern: "^/"
    auth:
      cookie:
        name: "SESSION_ID"
      jwt:
        secret: "your-32-byte-secret-key-here!!!!"
        expiry: 3600
      bearer:
        enabled: true

access:
  - path: /login
    require: NO_AUTH
  - path: /admin
    require: ROLE_ADMIN
  - path: /
    require: AUTH_REQUIRED

session:
  cookie: "SESSION_ID"
  cookie_path: "/"
  expires: 3600
  secure: true
  allow_reclaim: false
  disable_subdomain_persistence: true

csrf:
  enabled: true
  secret: "your-32-byte-auth-key-here!!!!"
  secure: true
  same_site: "Lax"
  field_name: "csrf.token"
  header_name: "X-CSRF-Token"

logout:
  enabled: true
  logout_url: "/logout"
  delete_cookies: ["SESSION_ID"]
  redirect_url: "/login"

entry_point:
  login_url: "/login"
  code: 302

access_denied:
  enabled: true
  url: "/access-denied"
`)

	config, err := LoadSecurityConfigFromBytes(yaml)
	if err != nil {
		t.Fatal(err)
	}

	if len(config.Firewalls) != 1 {
		t.Errorf("expected 1 firewall, got %d", len(config.Firewalls))
	}

	fw := config.Firewalls["main"]
	if fw.Auth.Cookie == nil {
		t.Error("expected cookie auth")
	}
	if fw.Auth.Cookie.Name != "SESSION_ID" {
		t.Errorf("expected cookie name SESSION_ID, got %s", fw.Auth.Cookie.Name)
	}
	if fw.Auth.JWT == nil {
		t.Error("expected JWT auth")
	}
	if fw.Auth.Bearer == nil || !fw.Auth.Bearer.Enabled {
		t.Error("expected bearer auth enabled")
	}

	if len(config.Access) != 3 {
		t.Errorf("expected 3 access rules, got %d", len(config.Access))
	}

	if config.Session == nil {
		t.Error("expected session config")
	}
	if config.Session.Cookie != "SESSION_ID" {
		t.Errorf("expected cookie SESSION_ID, got %s", config.Session.Cookie)
	}

	if config.CSRF == nil {
		t.Error("expected CSRF config")
	}
	if !config.CSRF.Enabled {
		t.Error("expected CSRF enabled")
	}

	if config.Logout == nil {
		t.Error("expected logout config")
	}
	if config.Logout.LogoutUrl != "/logout" {
		t.Errorf("expected logout_url /logout, got %s", config.Logout.LogoutUrl)
	}

	if config.EntryPoint == nil {
		t.Error("expected entry_point config")
	}
	if config.EntryPoint.LoginUrl != "/login" {
		t.Errorf("expected login_url /login, got %s", config.EntryPoint.LoginUrl)
	}

	if config.AccessDenied == nil {
		t.Error("expected access_denied config")
	}
	if config.AccessDenied.Url != "/access-denied" {
		t.Errorf("expected url /access-denied, got %s", config.AccessDenied.Url)
	}
}

func TestLoadRoutesConfig(t *testing.T) {
	yaml := []byte(`
routes:
  - path: /
    method: [GET]
    handler: Home.index
    require: AUTH_REQUIRED

  - path: /login
    handlers:
      GET: Home.login
      POST: Home.loginPost
    require: NO_AUTH

  - path: /api/data
    method: [POST]
    handler: Api.data
`)

	routes, err := LoadRoutesConfigFromBytes(yaml)
	if err != nil {
		t.Fatal(err)
	}

	if len(routes) != 3 {
		t.Errorf("expected 3 routes, got %d", len(routes))
	}

	route1 := routes[0]
	if route1.Path != "/" {
		t.Errorf("expected path /, got %s", route1.Path)
	}
	if len(route1.Method) != 1 || route1.Method[0] != "GET" {
		t.Errorf("expected method [GET], got %v", route1.Method)
	}
	if route1.Handler != "Home.index" {
		t.Errorf("expected handler Home.index, got %s", route1.Handler)
	}
	if route1.Require != "AUTH_REQUIRED" {
		t.Errorf("expected require AUTH_REQUIRED, got %s", route1.Require)
	}

	route2 := routes[1]
	if route2.Path != "/login" {
		t.Errorf("expected path /login, got %s", route2.Path)
	}
	if route2.Handlers == nil {
		t.Error("expected handlers map")
	}
	if route2.Handlers["GET"] != "Home.login" {
		t.Errorf("expected GET handler Home.login, got %s", route2.Handlers["GET"])
	}
	if route2.Handlers["POST"] != "Home.loginPost" {
		t.Errorf("expected POST handler Home.loginPost, got %s", route2.Handlers["POST"])
	}
	if route2.Require != "NO_AUTH" {
		t.Errorf("expected require NO_AUTH, got %s", route2.Require)
	}
}

func TestNewSecurityFromConfig(t *testing.T) {
	yaml := []byte(`
firewalls:
  main:
    pattern: "^/"
    auth:
      jwt:
        secret: "your-32-byte-secret-key-here!!!!"
        expiry: 3600

access:
  - path: /login
    require: NO_AUTH

session:
  cookie: "SESSION_ID"

csrf:
  enabled: false
`)

	config, err := LoadSecurityConfigFromBytes(yaml)
	if err != nil {
		t.Fatal(err)
	}

	sec, err := NewSecurityFromConfig(config)
	if err != nil {
		t.Fatal(err)
	}

	if sec == nil {
		t.Error("expected non-nil security")
	}

	if sec.Authenticator() == nil {
		t.Error("expected authenticator")
	}

	if sec.Session() == nil {
		t.Error("expected session manager")
	}
}

func TestUserClaims(t *testing.T) {
	claims := user.Claims{
		"sub":      "user123",
		"username": "john",
		"roles":    []string{"admin", "editor"},
	}

	if claims.GetID() != "user123" {
		t.Errorf("expected id user123, got %v", claims.GetID())
	}
	if claims.GetUsername() != "john" {
		t.Errorf("expected username john, got %s", claims.GetUsername())
	}
	if !claims.HasRole("admin") {
		t.Error("expected HasRole admin=true")
	}
	if claims.HasRole("guest") {
		t.Error("expected HasRole guest=false")
	}
}

func TestRoleCheckerFunc(t *testing.T) {
	checker := user.RoleCheckerFunc(func(ctx iris.Context, u user.Interface, role string) bool {
		return u.HasRole(role)
	})

	claims := user.Claims{
		"sub":      "user123",
		"username": "john",
		"roles":    []string{"admin"},
	}

	if !checker.CheckRole(nil, claims, "admin") {
		t.Error("expected role check true")
	}
	if checker.CheckRole(nil, claims, "guest") {
		t.Error("expected role check false")
	}
}
