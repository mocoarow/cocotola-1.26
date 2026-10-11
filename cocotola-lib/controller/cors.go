package controller

import (
	"errors"
	"slices"

	"github.com/gin-contrib/cors"
)

const wildcardOrigin = "*"

// ErrCORSCredentialsWithWildcard is returned when AllowCredentials is true but AllowOrigins is wildcard.
var ErrCORSCredentialsWithWildcard = errors.New("CORS: AllowCredentials=true with wildcard origin is not allowed; set specific AllowOrigins")

// ErrCORSWildcardMixedWithOrigins is returned when AllowOrigins contains the wildcard together with any other entry.
var ErrCORSWildcardMixedWithOrigins = errors.New("CORS: wildcard origin '*' must be the only entry in AllowOrigins")

// CORSConfig holds allowed origins, methods, headers, and credentials as comma-separated strings.
type CORSConfig struct {
	AllowOrigins     string `yaml:"allowOrigins" validate:"required"`
	AllowMethods     string `yaml:"allowMethods" validate:"required"`
	AllowHeaders     string `yaml:"allowHeaders" validate:"required"`
	AllowCredentials bool   `yaml:"allowCredentials"`
}

// InitCORS converts CORSConfig into a gin-contrib/cors.Config.
// Returns ErrCORSWildcardMixedWithOrigins when the wildcard origin is listed with other entries,
// and ErrCORSCredentialsWithWildcard when AllowCredentials is true with a wildcard origin.
func InitCORS(cfg *CORSConfig) (cors.Config, error) {
	allowOrigins := SplitCommaSeparated(cfg.AllowOrigins)
	allowMethods := SplitCommaSeparated(cfg.AllowMethods)
	allowHeaders := SplitCommaSeparated(cfg.AllowHeaders)

	// gin-contrib/cors turns any "*" entry into allow-all, which would also skip the credentials check below.
	if len(allowOrigins) > 1 && slices.Contains(allowOrigins, wildcardOrigin) {
		return cors.Config{}, ErrCORSWildcardMixedWithOrigins
	}

	if len(allowOrigins) == 1 && allowOrigins[0] == wildcardOrigin {
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
