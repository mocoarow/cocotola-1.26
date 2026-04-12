package config

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"go.yaml.in/yaml/v4"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"

	libcontroller "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
	libgateway "github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
)

// SupabaseConfig holds the Supabase JWKS URL for token verification.
type SupabaseConfig struct {
	JWKSURL string `yaml:"jwksUrl" validate:"required,url"`
}

// AuthConfig holds JWT signing key, token TTL, cookie delivery settings, and API key for service-to-service authentication.
type AuthConfig struct {
	SigningKey         string                  `yaml:"signingKey" validate:"required,min=32"`
	AccessTokenTTLMin  int                     `yaml:"accessTokenTtlMin" validate:"gte=1"`
	RefreshTokenTTLMin int                     `yaml:"refreshTokenTtlMin" validate:"gte=1"`
	SessionTokenTTLMin int                     `yaml:"sessionTokenTtlMin" validate:"gte=1"`
	SessionMaxTTLMin   int                     `yaml:"sessionMaxTtlMin" validate:"gte=1"`
	TokenWhitelistSize int                     `yaml:"tokenWhitelistSize" validate:"gte=1"`
	Cookie             controller.CookieConfig `yaml:"cookie" validate:"required"`
	Supabase           SupabaseConfig          `yaml:"supabase" validate:"required"`
	APIKey             string                  `yaml:"apiKey" validate:"required,min=32"`
}

// Config holds all configuration for the cocotola-auth service.
type Config struct {
	Server libcontroller.ServerConfig `yaml:"server" validate:"required"`
	DB     libgateway.DBConfig        `yaml:"db" validate:"required"`
	Trace  libgateway.TraceConfig     `yaml:"trace" validate:"required"`
	Log    libgateway.LogConfig       `yaml:"log" validate:"required"`
	Auth   AuthConfig                 `yaml:"auth" validate:"required"`
}

//go:embed config.yml
var config embed.FS

// ExpandEnvWithDefaults expands environment variables in the format VAR_NAME:-default_value.
func ExpandEnvWithDefaults(varName string) string {
	// Check if it contains :-
	if strings.Contains(varName, ":-") {
		parts := strings.SplitN(varName, ":-", 2) //nolint:mnd // split into name and default
		name := parts[0]
		defaultValue := parts[1]

		if value := os.Getenv(name); value != "" {
			return value
		}

		return defaultValue
	}

	// Simple variable expansion
	return os.Getenv(varName)
}

// LoadConfig reads the embedded config.yml file, expands environment variables, and returns a validated Config.
func LoadConfig() (*Config, error) {
	filename := "config.yml"
	confContent, err := config.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("config.ReadFile. filename: %s, err: %w", filename, err)
	}

	confContent = []byte(os.Expand(string(confContent), ExpandEnvWithDefaults))
	var conf Config
	if err := yaml.Unmarshal(confContent, &conf); err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal. filename: %s, err: %w", filename, err)
	}

	if err := domain.ValidateStruct(&conf); err != nil {
		return nil, fmt.Errorf("validate struct. filename: %s, err: %w", filename, err)
	}

	return &conf, nil
}
