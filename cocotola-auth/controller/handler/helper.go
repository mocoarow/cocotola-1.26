package handler

import (
	"fmt"
	"math"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// GetAppUserIDFromContext extracts a domain.AppUserID from the Gin context or
// returns a zero value + false if not set. Middleware stores the VO directly.
func GetAppUserIDFromContext(c *gin.Context) (domain.AppUserID, bool) {
	v, ok := c.Get(controller.ContextFieldUserID{})
	if !ok {
		return domain.AppUserID{}, false
	}
	id, ok := v.(domain.AppUserID)
	if !ok {
		return domain.AppUserID{}, false
	}
	return id, !id.IsZero()
}

// BridgeAppUserIDToInt32 returns -1 as an int32 placeholder because OpenAPI
// types still use int32 IDs while the domain uses UUIDv7.
// TODO(uuidv7-phase1-openapi): replace once the OpenAPI schema migrates to string IDs.
func BridgeAppUserIDToInt32(_ domain.AppUserID) (int32, error) {
	return -1, nil
}

// BridgeOrganizationIDToInt32 returns -1 as an int32 placeholder because OpenAPI
// types still use int32 IDs while the domain uses UUIDv7.
// TODO(uuidv7-phase1-openapi): replace once the OpenAPI schema migrates to string IDs.
func BridgeOrganizationIDToInt32(_ domain.OrganizationID) (int32, error) {
	return -1, nil
}

// GetIntFromPath extracts an integer value from the URL path parameter with the given name.
func GetIntFromPath(c *gin.Context, param string) (int, error) {
	idS := c.Param(param)
	id, err := strconv.Atoi(idS)
	if err != nil {
		return 0, fmt.Errorf("convert string to int(%s): %w", idS, err)
	}

	return id, nil
}

// SafeIntToInt32 converts an int to int32 with overflow check.
func SafeIntToInt32(v int) (int32, error) {
	if v < math.MinInt32 || v > math.MaxInt32 {
		return 0, fmt.Errorf("value %d overflows int32", v)
	}
	return int32(v), nil
}
