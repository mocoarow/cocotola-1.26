// Package config provides configuration loading for the cocotola-init application.
package config

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"go.yaml.in/yaml/v4"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"

	libgateway "github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
)

// InitConfig holds the first owner's login credentials.
type InitConfig struct {
	OwnerLoginID  string `yaml:"ownerLoginId" validate:"required"`
	OwnerPassword string `yaml:"ownerPassword" validate:"required,min=8"`
}

// Config holds all configuration for the cocotola-init application.
type Config struct {
	App InitConfig           `yaml:"app" validate:"required"`
	DB  libgateway.DBConfig  `yaml:"db" validate:"required"`
	Log libgateway.LogConfig `yaml:"log" validate:"required"`
}

//go:embed config.yml
var config embed.FS

const envVarSplitParts = 2

// expandEnvWithDefaults expands environment variables in the format VAR_NAME:-default_value.
func expandEnvWithDefaults(varName string) string {
	if strings.Contains(varName, ":-") {
		parts := strings.SplitN(varName, ":-", envVarSplitParts)
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
	confContent, err := config.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("config.ReadFile. filename: %s, err: %w", filename, err)
	}

	confContent = []byte(os.Expand(string(confContent), expandEnvWithDefaults))
	var conf Config
	if err := yaml.Unmarshal(confContent, &conf); err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal. filename: %s, err: %w", filename, err)
	}

	if err := domain.ValidateStruct(&conf); err != nil {
		return nil, fmt.Errorf("validate struct. filename: %s, err: %w", filename, err)
	}

	return &conf, nil
}
