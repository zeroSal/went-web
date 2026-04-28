package security

import (
	"embed"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// SecurityConfig holds the main security configuration
type SecurityConfig struct {
	Firewalls    map[string]FirewallConfig `yaml:"firewalls"`
	Access       []AccessRule              `yaml:"access"`
	Session      *SessionConfig            `yaml:"session"`
	CSRF         *CSRFConfig              `yaml:"csrf"`
	Logout       *LogoutConfig             `yaml:"logout"`
	EntryPoint   *EntryPointConfig        `yaml:"entry_point"`
	AccessDenied *AccessDeniedConfig      `yaml:"access_denied"`
}

type LogoutConfig struct {
	Enabled       bool     `yaml:"enabled"`
	LogoutUrl     string   `yaml:"logout_url"`
	DeleteCookies []string `yaml:"delete_cookies"`
	RedirectUrl   string   `yaml:"redirect_url"`
}

type EntryPointConfig struct {
	LoginUrl string `yaml:"login_url"`
	Code     int    `yaml:"code"`
}

type AccessDeniedConfig struct {
	Enabled bool   `yaml:"enabled"`
	Url     string `yaml:"url"`
}

type SessionConfig struct {
	Cookie                      string `yaml:"cookie"`
	CookiePath                  string `yaml:"cookie_path"`
	Domain                      string `yaml:"domain"`
	Expires                     int    `yaml:"expires"`
	Secure                      bool   `yaml:"secure"`
	AllowReclaim                bool   `yaml:"allow_reclaim"`
	DisableSubdomainPersistence bool   `yaml:"disable_subdomain_persistence"`
}

type CSRFConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Secret     string `yaml:"secret"`
	Secure     bool   `yaml:"secure"`
	SameSite   string `yaml:"same_site"`
	FieldName  string `yaml:"field_name"`
	HeaderName string `yaml:"header_name"`
}

type AccessRule struct {
	Path    string `yaml:"path"`
	Require string `yaml:"require"`
}

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

// Duration wraps an integer to handle YAML unmarshaling
type Duration struct {
	Duration int `yaml:"duration"`
}

// UnmarshalYAML handles both direct integer and structured duration formats
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var i int
	if err := unmarshal(&i); err == nil {
		d.Duration = i
		return nil
	}
	var s struct {
		Duration int `yaml:"duration"`
	}
	if err := unmarshal(&s); err != nil {
		return err
	}
	d.Duration = s.Duration
	return nil
}

// ToDuration converts to time.Duration (seconds)
func (d Duration) ToDuration() time.Duration {
	return time.Duration(d.Duration)
}

// RouteConfig defines a single route configuration
type RouteConfig struct {
	Path     string            `yaml:"path"`
	Method   []string          `yaml:"method"`
	Handler  string            `yaml:"handler"`
	Handlers map[string]string `yaml:"handlers"`
	Require  string            `yaml:"require"`
}

// RoutesConfig holds a list of route configurations
type RoutesConfig struct {
	Routes []RouteConfig `yaml:"routes"`
}

// LoadSecurityConfig loads security config from a file path
func LoadSecurityConfig(path string) (*SecurityConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadSecurityConfigFromBytes(data)
}

// LoadSecurityConfigFromBytes loads security config from YAML bytes
func LoadSecurityConfigFromBytes(data []byte) (*SecurityConfig, error) {
	var config SecurityConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadSecurityConfigFromEmbed loads security config from an embedded filesystem
func LoadSecurityConfigFromEmbed(efs embed.FS, path string) (*SecurityConfig, error) {
	data, err := efs.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadSecurityConfigFromBytes(data)
}

// LoadRoutesConfig loads routes config from a file path
func LoadRoutesConfig(path string) ([]RouteConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadRoutesConfigFromBytes(data)
}

// LoadRoutesConfigFromBytes loads routes config from YAML bytes
func LoadRoutesConfigFromBytes(data []byte) ([]RouteConfig, error) {
	var config RoutesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return config.Routes, nil
}

// LoadRoutesConfigFromEmbed loads routes config from an embedded filesystem
func LoadRoutesConfigFromEmbed(efs embed.FS, path string) ([]RouteConfig, error) {
	data, err := efs.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadRoutesConfigFromBytes(data)
}
