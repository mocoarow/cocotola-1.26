// Package config loads configuration for cocotola-audio-generator.
package config

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"go.yaml.in/yaml/v4"
)

// Config is the root configuration.
type Config struct {
	AppEnv string               `yaml:"appEnv" validate:"required"`
	Log    LogConfig            `yaml:"log"`
	Audio  AudioGeneratorConfig `yaml:"audio" validate:"required"`
}

// LogConfig is a thin logging config; the full LogConfig from cocotola-lib is
// not pulled in to keep this service's dependencies small.
type LogConfig struct {
	Level string `yaml:"level"`
}

// AudioGeneratorConfig collects all knobs that drive the batch.
type AudioGeneratorConfig struct {
	BucketName            string `yaml:"bucketName" validate:"required"`
	MaxPerRun             int    `yaml:"maxPerRun" validate:"required,gte=1,lte=1000"`
	TTSVoiceJa            string `yaml:"ttsVoiceJa" validate:"required"`
	TTSVoiceEn            string `yaml:"ttsVoiceEn" validate:"required"`
	AudioEncoding         string `yaml:"audioEncoding" validate:"required,oneof=OGG_OPUS MP3"`
	SampleRateHz          int    `yaml:"sampleRateHz" validate:"required,gte=8000,lte=48000"`
	QuestionAPIBaseURL    string `yaml:"questionApiBaseUrl" validate:"required,url"`
	QuestionAPIKey        string `yaml:"questionApiKey" validate:"required"`
	QuestionAPITimeoutSec int    `yaml:"questionApiTimeoutSec" validate:"required,gte=1,lte=120"`
	// StaleAfterSec is how long an audio entry may stay in "generating" before
	// the reaper at the start of each run reclaims it back to "pending".
	StaleAfterSec int `yaml:"staleAfterSec" validate:"required,gte=60,lte=86400"`
}

//go:embed config.yml
var configFS embed.FS

const envDefaultSeparator = ":-"

// expandEnvWithDefaults expands environment variables in the format VAR_NAME:-default_value.
func expandEnvWithDefaults(varName string) string {
	if strings.Contains(varName, envDefaultSeparator) {
		name, defaultValue, _ := strings.Cut(varName, envDefaultSeparator)
		if value := os.Getenv(name); value != "" {
			return value
		}
		return defaultValue
	}
	return os.Getenv(varName)
}

// LoadConfig reads the embedded config.yml, expands env vars, and validates.
func LoadConfig() (*Config, error) {
	raw, err := configFS.ReadFile("config.yml")
	if err != nil {
		return nil, fmt.Errorf("read config.yml: %w", err)
	}
	expanded := []byte(os.Expand(string(raw), expandEnvWithDefaults))
	var cfg Config
	if err := yaml.Unmarshal(expanded, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config.yml: %w", err)
	}
	if err := validator.New().Struct(&cfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	return &cfg, nil
}
