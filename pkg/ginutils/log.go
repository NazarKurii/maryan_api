package ginutil

import (
	"bytes"
	"encoding/json"
	"io"
	"maryan_api/pkg/log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func LogMiddlewear(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		queryParams, _ := json.Marshal(c.Request.URL.Query())
		headers, _ := json.Marshal(c.Request.Header)
		body, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		logger := log.New(c.ClientIP(), c.FullPath(), queryParams, headers, body, c.Request.Method)
		defer logger.Do(db)

		c.Set("logger", logger)
		c.Next()
	}
}

func getLogger(c *gin.Context) log.Logger {
	return c.MustGet("logger").(log.Logger)
}
