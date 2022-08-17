package config

import (
	"github.com/txsvc/stdlib/v2/env"
	"github.com/txsvc/stdlib/v2/settings"
)

const (
	defaultServiceEndpoint = "https://podops.dev"
	defaultAPIEndpoint     = "https://api.podops.dev"
	defaultContentEndpoint = "https://cdn.podops.dev"

	// ENV variables
	PodopsConfigPathEnv      = "PODOPS_CONFIG_PATH"
	PodopsStorageLocationEnv = "PODOPS_STORAGE_LOCATION"
	PodopsStaticLocationEnv  = "PODOPS_STATIC_LOCATION"
	PodopsBuildLocationEnv   = "PODOPS_BUILD_LOCATION"
	PodopsServiceEndpointEnv = "PODOPS_SERVICE_ENDPOINT"
	PodopsAPIEndpointEnv     = "PODOPS_API_ENDPOINT"
	PodopsContentEndpointEnv = "PODOPS_CONTENT_ENDPOINT"

	// default scopes
	ScopeContentAdmin  = "content:admin"
	ScopeContentEditor = "content:editor"
	ScopeContentRead   = "content:read"
	ScopeContentWrite  = "content:write"

	// other constants
	DefaultFeedName = "feed.xml"
	DefaultShowName = "show.yaml"
)

var (
	cfg        *settings.DialSettings
	configPath string
)

func init() {
	UpdateClientSettings("")
}

func Settings() *settings.DialSettings {
	return cfg
}

func SettingsPath() string {
	return configPath
}

// DefaultClientSettings returns a configuration based defaults and ENV variables.
func DefaultClientSettings() *settings.DialSettings {
	s := settings.DialSettings{
		Endpoint:        env.GetString(PodopsAPIEndpointEnv, defaultAPIEndpoint),
		DefaultEndpoint: defaultAPIEndpoint,
		UserAgent:       UserAgentString,
	}
	WithServiceEndpoint(env.GetString(PodopsServiceEndpointEnv, defaultServiceEndpoint)).Apply(&s)
	WithContentEndpoint(env.GetString(PodopsContentEndpointEnv, defaultContentEndpoint)).Apply(&s)

	return &s
}

// LoadClientSettings loads the configuration from file. Stored values will be overwritten
// from ENV values if these are set.
func LoadClientSettings(path string) (*settings.DialSettings, string) {
	var cfg *settings.DialSettings
	var p string

	paths := ResolveConfigPath(path)
	for _, path := range paths {
		c, err := settings.ReadSettingsFromFile(path)
		p = path
		if err == nil {
			cfg = c
			break
		}
	}
	if cfg == nil {
		return DefaultClientSettings(), DefaultConfigFileLocation
	}

	// overwrite with ENV if set
	if env.Exists(PodopsAPIEndpointEnv) {
		cfg.Endpoint = env.GetString(PodopsAPIEndpointEnv, "")
	}
	if env.Exists(PodopsServiceEndpointEnv) {
		WithServiceEndpoint(env.GetString(PodopsServiceEndpointEnv, "")).Apply(cfg)
	}
	if env.Exists(PodopsContentEndpointEnv) {
		WithContentEndpoint(env.GetString(PodopsContentEndpointEnv, "")).Apply(cfg)
	}
	return cfg, p
}

// UpdateClientSettings load the configuration from the given path and updates the global settings.
func UpdateClientSettings(path string) *settings.DialSettings {
	cfg, configPath = LoadClientSettings(path)
	return cfg
}

// WithServiceEndpoint returns a ClientOption that overrides the default service endpoint to be used for a client.
func WithServiceEndpoint(url string) settings.ClientOption {
	return withServiceEndpoint(url)
}

type withServiceEndpoint string

func (w withServiceEndpoint) Apply(o *settings.DialSettings) {
	o.SetOption(PodopsServiceEndpointEnv, string(w))
}

// WithContentEndpoint returns a ClientOption that overrides the default cdn endpoint to be used for a client.
func WithContentEndpoint(url string) settings.ClientOption {
	return withContentEndpoint(url)
}

type withContentEndpoint string

func (w withContentEndpoint) Apply(o *settings.DialSettings) {
	o.SetOption(PodopsContentEndpointEnv, string(w))
}
