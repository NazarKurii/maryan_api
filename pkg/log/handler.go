package log

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func RegisterLogRoute(db *gorm.DB, s *gin.Engine) {
	s.GET("/instances/:id", func(c *gin.Context) {

		idParam := c.Param("id")
		logID, err := uuid.Parse(idParam)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid UUID format")
			return
		}

		var log Log
		if err := db.First(&log, "id = ?", logID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.String(http.StatusNotFound, "Log not found")
			} else {
				c.String(http.StatusInternalServerError, "Database error")
			}
			return
		}

		htmlBytes, err := log.HTML()
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to render HTML")
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", htmlBytes)

	})
}
