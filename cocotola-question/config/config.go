package config

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"go.yaml.in/yaml/v4"

	libcontroller "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
	libgateway "github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// QuestionConfig holds configuration for the cocotola-question service.
type QuestionConfig struct {
	FirestoreProjectID string `yaml:"firestoreProjectId" validate:"required"`
}

// AuthClientConfig holds configuration for the auth service HTTP client.
type AuthClientConfig struct {
	BaseURL    string `yaml:"baseUrl" validate:"required"`
	Audience   string `yaml:"audience"`
	APIKey     string `yaml:"apiKey" validate:"required"`
	TimeoutSec int    `yaml:"timeoutSec" validate:"required,gte=1"`
}

// Config holds all configuration for the standalone cocotola-question service.
type Config struct {
	AppEnv   string                     `yaml:"appEnv" validate:"required"`
	Server   libcontroller.ServerConfig `yaml:"server" validate:"required"`
	Question QuestionConfig             `yaml:"question" validate:"required"`
	Auth     AuthClientConfig           `yaml:"auth" validate:"required"`
	Trace    libgateway.TraceConfig     `yaml:"trace" validate:"required"`
	Log      libgateway.LogConfig       `yaml:"log" validate:"required"`
}

//go:embed config.yml
var configFS embed.FS

const envDefaultSeparator = ":-"

// ExpandEnvWithDefaults expands environment variables in the format VAR_NAME:-default_value.
func ExpandEnvWithDefaults(varName string) string {
	if strings.Contains(varName, envDefaultSeparator) {
		name, defaultValue, _ := strings.Cut(varName, envDefaultSeparator)

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

	if err := domain.ValidateStruct(&conf); err != nil {
		return nil, fmt.Errorf("validate config(%s): %w", filename, err)
	}

	return &conf, nil
}
