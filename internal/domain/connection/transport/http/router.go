package http

import (
	"net/http"

	"github.com/nazarkurii/marshrutka_api/internal/domain/connection/repo"
	"github.com/nazarkurii/marshrutka_api/internal/domain/connection/service"
	"github.com/nazarkurii/marshrutka_api/pkg/auth"
	ginutil "github.com/nazarkurii/marshrutka_api/pkg/ginutils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"gorm.io/gorm"
)

func RegisterRoutes(db *gorm.DB, rdb *redis.Client, s *gin.Engine, client *http.Client) {
	adminRouter := ginutil.CreateAuthRouter("/admin", auth.Admin.SecretKey(), s)
	customerRouter := s.Group("/customer")
	adminHandler := newAdminHandler(service.NewAdminConnection(repo.NewConnectionRepo(db), repo.NewConnectionCacheRepo(rdb)))
	customerHandler := newCustomerHandler(service.NewCustomerConnection(repo.NewConnectionRepo(db), repo.NewConnectionCacheRepo(rdb)))

	//-----------------------Trip Routes---------------------------------------

	adminRouter.GET("/connection/:id", adminHandler.GetByID)
	adminRouter.GET("/connections", adminHandler.GetConnections)
	adminRouter.POST("/connection/update", adminHandler.RegisterUpdate)

	customerRouter.GET("/connection/:id", customerHandler.GetByID)
	customerRouter.GET("/connections", customerHandler.GetConnections)
	customerRouter.GET("/connections/:from/:to/:date/:adults/:children/:teenagers", customerHandler.FindConnections)
}
