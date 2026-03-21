package handler

import (
	"fmt"
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
