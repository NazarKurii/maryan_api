package http

import (
	"maryan_api/internal/domain/trip/repo"
	"maryan_api/internal/domain/trip/service"
	"maryan_api/pkg/auth"
	ginutil "maryan_api/pkg/ginutils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(db *gorm.DB, s *gin.Engine, client *http.Client) {
	adminRouter := ginutil.CreateAuthRouter("/admin", auth.Admin.SecretKey(), s)

	handler := newTripHandler(service.NewTripService(repo.NewTrip(db), repo.NewBus(db), repo.NewCountry(db)))
	//-----------------------Trip Routes---------------------------------------
	adminRouter.POST("/trip", handler.Create)
	adminRouter.POST("/trip/test", handler.CreateTest)
	adminRouter.GET("/trip/:id", handler.GetByID)
	adminRouter.GET("/trips", handler.GetTrips)
	adminRouter.POST("/trip/update", handler.RegisterUpdate)
}
