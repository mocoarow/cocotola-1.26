package handler

import (
	"fmt"
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
