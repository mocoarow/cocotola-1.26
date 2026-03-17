package handler

import (
	libcontroller "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
)

// LogConfig controls access log output settings.
type LogConfig struct {
	AccessLog             bool `yaml:"accessLog"`
	AccessLogRequestBody  bool `yaml:"accessLogRequestBody"`
	AccessLogResponseBody bool `yaml:"accessLogResponseBody"`
}

// DebugConfig controls debug-mode features (Gin debug mode, artificial wait).
type DebugConfig struct {
	Gin  bool `yaml:"gin"`
	Wait bool `yaml:"wait"`
}

// Config holds handler-level configuration for CORS, logging, and debug settings.
type Config struct {
	CORS  libcontroller.CORSConfig `yaml:"cors" validate:"required"`
	Log   LogConfig                `yaml:"log" validate:"required"`
	Debug DebugConfig              `yaml:"debug" validate:"required"`
}
