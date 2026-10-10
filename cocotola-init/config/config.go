// Package config provides configuration loading for the cocotola-init application.
package config

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"go.yaml.in/yaml/v4"

	libdomain "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain"
	libgateway "github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
)

// InitConfig holds the first owner's login credentials.
type InitConfig struct {
	OwnerLoginID  string `yaml:"ownerLoginId" validate:"required"`
	OwnerPassword string `yaml:"ownerPassword" validate:"required,min=8"`
}

// QuestionClientConfig holds connection settings for the cocotola-question
// internal API. TimeoutSec 0 means the built-in default.
type QuestionClientConfig struct {
	BaseURL    string `yaml:"baseUrl"`
	APIKey     string `yaml:"apiKey" validate:"required_with=BaseURL"`
	TimeoutSec int    `yaml:"timeoutSec" validate:"gte=0"`
}

// CSVSeedConfig holds the GCS bucket that the import mode reads CSV objects from.
type CSVSeedConfig struct {
	BucketName string `yaml:"bucketName"`
}

// Config holds all configuration for the cocotola-init application. Settings
// required by only one mode are checked by ValidateForMode.
type Config struct {
	AppEnv   string                 `yaml:"appEnv" validate:"required"`
	App      InitConfig             `yaml:"app" validate:"-"`
	DB       libgateway.DBConfig    `yaml:"db" validate:"required"`
	Question QuestionClientConfig   `yaml:"question"`
	CSVSeed  CSVSeedConfig          `yaml:"csvSeed"`
	Log      libgateway.LogConfig   `yaml:"log" validate:"required"`
	Trace    libgateway.TraceConfig `yaml:"trace" validate:"required"`
}

// Mode selects what cocotola-init does.
type Mode int

const (
	_ Mode = iota
	// ModeInit bootstraps the organization, users and embedded public workbooks.
	ModeInit
	// ModeImport adds the rows appended to the CSV workbooks.
	ModeImport
)

var (
	// ErrUnknownMode is returned for a command-line mode other than init or import.
	ErrUnknownMode = errors.New("unknown mode")
	// ErrInvalidModeConfig is returned when a setting the mode needs is missing.
	ErrInvalidModeConfig = errors.New("invalid config for mode")
)

// ParseMode returns the mode named by the first command-line argument; no
// argument means ModeInit.
func ParseMode(args []string) (Mode, error) {
	if len(args) == 0 {
		return ModeInit, nil
	}
	if len(args) > 1 {
		return 0, fmt.Errorf("args %q: %w", args, ErrUnknownMode)
	}

	switch args[0] {
	case "init":
		return ModeInit, nil
	case "import":
		return ModeImport, nil
	default:
		return 0, fmt.Errorf("mode %q: %w", args[0], ErrUnknownMode)
	}
}

// ValidateForMode checks the settings that LoadConfig leaves to the mode.
func (c *Config) ValidateForMode(mode Mode) error {
	if mode != ModeInit && mode != ModeImport {
		return fmt.Errorf("mode %d: %w", mode, ErrUnknownMode)
	}
	if c.Question.BaseURL == "" {
		return fmt.Errorf("question.baseUrl: %w", ErrInvalidModeConfig)
	}

	if mode == ModeInit {
		if err := libdomain.ValidateStruct(&c.App); err != nil {
			return fmt.Errorf("app: %w: %w", ErrInvalidModeConfig, err)
		}
		return nil
	}
	if c.CSVSeed.BucketName == "" {
		return fmt.Errorf("csvSeed.bucketName: %w", ErrInvalidModeConfig)
	}
	return nil
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

	validated := withoutUnusedGoogleTrace(conf)
	if err := libdomain.ValidateStruct(&validated); err != nil {
		return nil, fmt.Errorf("validate struct. filename: %s, err: %w", filename, err)
	}

	return &validated, nil
}

func withoutUnusedGoogleTrace(conf Config) Config {
	if conf.Trace.Exporter == "google" {
		return conf
	}
	trace := conf.Trace
	trace.Google = nil
	conf.Trace = trace
	return conf
}
