package router

import (
	"maryan_api/internal/domains/trip"
	"maryan_api/internal/domains/user"
	ginutil "maryan_api/pkg/ginutils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(s *gin.Engine, c *http.Client, db *gorm.DB) {
	s.Static("/imgs", "../../static/imgs")

	s.Use(ginutil.LogMiddlewear(db))
	user.RegisterRoutes(db, s, c)
	trip.RegisterRoutes(db, s)

}
