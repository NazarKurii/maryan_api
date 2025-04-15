package router

import (
	"maryan_api/internal/domains/user"
	"maryan_api/pkg/log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(s *gin.Engine, c *http.Client, db *gorm.DB) {
	user.RegisterRoutes(db, s, c)
	log.RegisterLogRoute(db, s)
}
