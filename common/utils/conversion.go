package utils

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ConvertQueryParamToInt(c *gin.Context, paramName string, defaultValue int, minValue int, maxValue int) (int, error) {
	paramString := c.DefaultQuery(paramName, strconv.Itoa(defaultValue))

	// Convert string to integer
	value, err := strconv.Atoi(paramString)
	if err != nil {
		return 0, err
	}

	// Check if the value is within the acceptable range
	if value < minValue || value > maxValue {
		return 0, fmt.Errorf("parameter %s is out of range", paramName)
	}

	return value, nil
}
