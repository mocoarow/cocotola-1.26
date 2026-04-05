// Package config provides configuration loading for the cocotola-app.
package config

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"go.yaml.in/yaml/v4"

	authconfig "github.com/mocoarow/cocotola-1.26/cocotola-auth/config"
	questionconfig "github.com/mocoarow/cocotola-1.26/cocotola-question/config"

	libcontroller "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
	libgateway "github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
)

// AppConfig holds application-level configuration for included microservices.
type AppConfig struct {
	Auth     authconfig.AuthConfig         `yaml:"auth" validate:"required"`
	Internal authconfig.InternalConfig     `yaml:"internal" validate:"required"`
	Question questionconfig.QuestionConfig `yaml:"question" validate:"required"`
}

// Config holds the complete application configuration.
type Config struct {
	App    AppConfig                  `yaml:"app" validate:"required"`
	Server libcontroller.ServerConfig `yaml:"server" validate:"required"`
	DB     libgateway.DBConfig        `yaml:"db" validate:"required"`
	Trace  libgateway.TraceConfig     `yaml:"trace" validate:"required"`
	Log    libgateway.LogConfig       `yaml:"log" validate:"required"`
}

//go:embed config.yml
var configFS embed.FS

// ExpandEnvWithDefaults expands environment variables in the format VAR_NAME:-default_value.
func ExpandEnvWithDefaults(varName string) string {
	if strings.Contains(varName, ":-") {
		parts := strings.SplitN(varName, ":-", 2) //nolint:mnd // split into name and default
		name := parts[0]
		defaultValue := parts[1]

		if value := os.Getenv(name); value != "" {
			return value
		}

		return defaultValue
	}

	return os.Getenv(varName)
}

// LoadConfig reads the embedded config.yml file, expands environment variables, and returns a validated Config.
func LoadConfig() (*Config, error) {
	filename := "config.yml"
	confContent, err := configFS.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read config file(%s): %w", filename, err)
	}

	confContent = []byte(os.Expand(string(confContent), ExpandEnvWithDefaults))

	var conf Config
	if err := yaml.Unmarshal(confContent, &conf); err != nil {
		return nil, fmt.Errorf("unmarshal file(%s): %w", filename, err)
	}

	return &conf, nil
}
