package controller

import (
	"errors"

	"github.com/gin-contrib/cors"
)

// ErrCORSCredentialsWithWildcard is returned when AllowCredentials is true but AllowOrigins is wildcard.
var ErrCORSCredentialsWithWildcard = errors.New("CORS: AllowCredentials=true with wildcard origin is not allowed; set specific AllowOrigins")

// CORSConfig holds allowed origins, methods, headers, and credentials as comma-separated strings.
type CORSConfig struct {
	AllowOrigins     string `yaml:"allowOrigins" validate:"required"`
	AllowMethods     string `yaml:"allowMethods" validate:"required"`
	AllowHeaders     string `yaml:"allowHeaders" validate:"required"`
	AllowCredentials bool   `yaml:"allowCredentials"`
}

// InitCORS converts CORSConfig into a gin-contrib/cors.Config.
// Returns ErrCORSCredentialsWithWildcard when AllowCredentials is true with a wildcard origin.
func InitCORS(cfg *CORSConfig) (cors.Config, error) {
	allowOrigins := SplitCommaSeparated(cfg.AllowOrigins)
	allowMethods := SplitCommaSeparated(cfg.AllowMethods)
	allowHeaders := SplitCommaSeparated(cfg.AllowHeaders)

	if len(allowOrigins) == 1 && allowOrigins[0] == "*" {
		if cfg.AllowCredentials {
			return cors.Config{}, ErrCORSCredentialsWithWildcard
		}

		return cors.Config{ //nolint:exhaustruct
			AllowAllOrigins: true,
			AllowMethods:    allowMethods,
			AllowHeaders:    allowHeaders,
		}, nil
	}

	return cors.Config{ //nolint:exhaustruct
		AllowOrigins:     allowOrigins,
		AllowMethods:     allowMethods,
		AllowHeaders:     allowHeaders,
		AllowCredentials: cfg.AllowCredentials,
	}, nil
}
