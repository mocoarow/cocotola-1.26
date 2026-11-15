package controller

import "time"

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

// ShutdownConfig holds graceful shutdown timeout settings.
type ShutdownConfig struct {
	GracePeriodSec  int `yaml:"gracePeriodSec" validate:"gte=1"`
	ShutdownTimeSec int `yaml:"shutdownTimeSec" validate:"gte=1"`
}

// ServerConfig holds HTTP server port, CORS, logging, debug, and shutdown settings.
type ServerConfig struct {
	HTTPPort             int            `yaml:"httpPort" validate:"required"`
	MetricsPort          int            `yaml:"metricsPort" validate:"required"`
	ReadHeaderTimeoutSec int            `yaml:"readHeaderTimeoutSec" validate:"gte=1"`
	ReadTimeoutSec       int            `yaml:"readTimeoutSec" validate:"gte=1"`
	WriteTimeoutSec      int            `yaml:"writeTimeoutSec" validate:"gte=1"`
	IdleTimeoutSec       int            `yaml:"idleTimeoutSec" validate:"gte=1"`
	CORS                 CORSConfig     `yaml:"cors" validate:"required"`
	Log                  LogConfig      `yaml:"log" validate:"required"`
	Debug                DebugConfig    `yaml:"debug" validate:"required"`
	Shutdown             ShutdownConfig `yaml:"shutdown" validate:"required"`
}

// HTTPServerTimeouts returns the timeouts for WebServerProcess and MetricsServerProcess.
func (c ServerConfig) HTTPServerTimeouts() HTTPServerTimeouts {
	return HTTPServerTimeouts{
		ReadHeader: time.Duration(c.ReadHeaderTimeoutSec) * time.Second,
		Read:       time.Duration(c.ReadTimeoutSec) * time.Second,
		Write:      time.Duration(c.WriteTimeoutSec) * time.Second,
		Idle:       time.Duration(c.IdleTimeoutSec) * time.Second,
		Shutdown:   time.Duration(c.Shutdown.ShutdownTimeSec) * time.Second,
	}
}
