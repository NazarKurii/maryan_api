package http

import (
	"net/http"

	"github.com/nazarkurii/marshrutka_api/internal/domain/tickets/repo"
	"github.com/nazarkurii/marshrutka_api/internal/domain/tickets/service"
	"github.com/nazarkurii/marshrutka_api/pkg/auth"
	ginutil "github.com/nazarkurii/marshrutka_api/pkg/ginutils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(db *gorm.DB, s *gin.Engine, client *http.Client) {
	customerRouter := ginutil.CreateAuthRouter("/customer", auth.Customer.SecretKey(), s)

	customerHandler := newHandler(service.NewTicketService(repo.NewTicketRepo(db), client))

	//-----------------------Ticket Routes---------------------------------------

	customerRouter.POST("/connection/purchase-ticket", customerHandler.purchase)
	customerRouter.GET("/tickets", customerHandler.getTickets)
	s.GET("/connection/purchase-ticket/failed/:id/:token", customerHandler.purchaseFailed)
	s.GET("/connection/purchase-ticket/succeded/:id/:token", customerHandler.purchaseSucceded)
}
