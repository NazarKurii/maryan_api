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

		var body []byte
		contentType := c.ContentType()

		if contentType != "multipart/form-data" {
			// Safe to read the body for non-multipart types
			body, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		} else {
			// For multipart, just log the form fields (not files)
			if err := c.Request.ParseMultipartForm(32 << 20); err == nil && c.Request.MultipartForm != nil {
				formData := make(map[string][]string)
				for k, v := range c.Request.MultipartForm.Value {
					formData[k] = v
				}
				body, _ = json.Marshal(formData)
			}
		}

		logger := log.New(
			c.ClientIP(),
			c.FullPath(),
			queryParams,
			headers,
			body,
			c.Request.Method,
		)
		defer logger.Do(db)

		c.Set("logger", logger)
		c.Next()
	}
}

func getLogger(c *gin.Context) log.Logger {
	return c.MustGet("logger").(log.Logger)
}
