package ginutil

import (
	"maryan_api/pkg/log"

	"github.com/gin-gonic/gin"
)

func RequestLog(c *gin.Context) log.Logger {
	return c.MustGet("log").(log.Logger)
}
