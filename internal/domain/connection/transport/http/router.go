package http

import (
	"maryan_api/internal/domain/connection/repo"
	"maryan_api/internal/domain/connection/service"
	"maryan_api/pkg/auth"
	ginutil "maryan_api/pkg/ginutils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(db *gorm.DB, s *gin.Engine, client *http.Client) {
	adminRouter := ginutil.CreateAuthRouter("/admin", auth.Admin.SecretKey(), s)
	customerRouter := ginutil.CreateAuthRouter("/customer", auth.Customer.SecretKey(), s)

	adminHandler := newAdminHandler(service.NewAdminConnection(repo.NewConnectionRepo(db)))
	customerHandler := newCustomerHandler(service.NewCustomerConnection(repo.NewConnectionRepo(db)))

	//-----------------------Trip Routes---------------------------------------

	adminRouter.GET("/connection/:id", adminHandler.GetByID)
	adminRouter.GET("/connections", adminHandler.GetConnections)
	adminRouter.POST("/connection/update", adminHandler.RegisterUpdate)

	customerRouter.GET("/connection/:id", customerHandler.GetByID)
	customerRouter.GET("/connections", customerHandler.GetConnections)
	customerRouter.GET("/connections/:from/:to/:date/:adults/:children/:teenagers", customerHandler.FindConnections)
}
