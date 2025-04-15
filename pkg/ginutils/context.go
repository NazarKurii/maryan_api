package ginutil

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParseIntKey(c *gin.Context, key string) (int, error) {
	value := c.Query(key)
	if value == "" {
		return 0, fmt.Errorf(`"%s" query param could not be found in the request`, key)
	}

	intValue, err := strconv.Atoi(value)

	if err != nil {
		return 0, fmt.Errorf(`Problem parsing "%s" query param: %s`, key, err.Error())
	}

	return intValue, nil
}

func ParseStringKey(c *gin.Context, key string) (string, error) {
	value := c.Query(key)
	if value == "" {
		return "", fmt.Errorf(`"%s" query param could not be found in the request`, key)
	}
	return value, nil
}
