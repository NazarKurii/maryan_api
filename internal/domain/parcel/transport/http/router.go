package http

import (
	"net/http"

	"github.com/nazarkurii/marshrutka_api/internal/domain/parcel/repo"
	"github.com/nazarkurii/marshrutka_api/internal/domain/parcel/service"
	"github.com/nazarkurii/marshrutka_api/pkg/auth"
	ginutil "github.com/nazarkurii/marshrutka_api/pkg/ginutils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(db *gorm.DB, s *gin.Engine, client *http.Client) {
	customerRouter := ginutil.CreateAuthRouter("/customer", auth.Customer.SecretKey(), s)

	customerHandler := newHandler(service.NewParcelService(repo.NewParcelRepo(db), client))

	s.GET("/connection/available-parcel-dates/:from/:to/:year/:month/:width/:height/:length", customerHandler.findConnections)
	customerRouter.POST("/connection/:id/purchase-parcel", customerHandler.purchase)
	customerRouter.GET("/connection-parcel/:id/:width/:height/:length", customerHandler.GetByID)
	customerRouter.GET("/parcels", customerHandler.getParcels)
	s.GET("/connection/purchase-parcel/failed/:id/:token", customerHandler.purchaseFailed)
	s.GET("/connection/purchase-parcel/succeded/:id/:token", customerHandler.purchaseSucceded)
}
