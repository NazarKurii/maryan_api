package router

import (
	bus "maryan_api/internal/domain/bus/transport/http"
	passenger "maryan_api/internal/domain/passenger/transport/http"
	user "maryan_api/internal/domain/user/transport/http"
	ginutil "maryan_api/pkg/ginutils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(s *gin.Engine, db *gorm.DB, client *http.Client) {
	s.Use(ginutil.LogMiddlewear(db))

	passenger.RegisterRoutes(db, s, client)
	user.RegisterRoutes(db, s, client)
	bus.RegisterRoutes(db, s, client)
}
