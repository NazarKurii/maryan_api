package http

import (
	"context"
	"maryan_api/config"
	"maryan_api/internal/domain/tickets/repo"
	"maryan_api/internal/domain/tickets/service"
	"maryan_api/internal/entity"
	"maryan_api/pkg/auth"
	"maryan_api/pkg/dbutil"
	ginutil "maryan_api/pkg/ginutils"
	"maryan_api/pkg/hypermedia"
	rfc7807 "maryan_api/pkg/problem"
	"net/http"
	"time"

	"github.com/d3code/uuid"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(db *gorm.DB, s *gin.Engine, client *http.Client) {
	customerRouter := ginutil.CreateAuthRouter("/customer", auth.Customer.SecretKey(), s)

	customerHandler := newHandler(service.NewTicketService(repo.NewTicketRepo(db)))

	//-----------------------Ticket Routes---------------------------------------

	customerRouter.POST("/connection/purchase-ticket", customerHandler.purchase)
	customerRouter.GET("/tickets", customerHandler.getTickets)
	s.GET("/connection/purchase-ticket/failed/:id", customerHandler.purchaseFailed)
	s.GET("/connection/purchase-ticket/succeded/:id", customerHandler.purchaseSucceded)
}

type passengerHandler struct {
	service service.Ticket
}

func newHandler(service service.Ticket) *passengerHandler {
	return &passengerHandler{service}
}

func (p *passengerHandler) getTickets(ctx *gin.Context) {

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	tickets, links, err := p.service.GetTickets(ctxWithTimeout, dbutil.PaginationStr{
		"/customer/tickets",
		ctx.DefaultQuery("page", "1"),
		ctx.DefaultQuery("size", "9"),
		ctx.DefaultQuery("order_by", "created_at"),
		ctx.DefaultQuery("order_way", "desc"),
		ctx.DefaultQuery("search", ""),
	}, ctx.MustGet("userID").(uuid.UUID))

	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		Links   hypermedia.Links        `json:"links"`
		Tickets []entity.CustomerTicket `json:"tickets"`
	}{
		ginutil.Response{
			"The adresses have successfuly beeen found.",
			hypermedia.Links{},
		},
		links,
		tickets,
	})

}

func (p *passengerHandler) purchase(ctx *gin.Context) {
	var request entity.NewTicketJSON

	err := ctx.ShouldBindJSON(&request)
	if err != nil {

		ginutil.HandlerProblemAbort(ctx, rfc7807.JSON(err.Error()))
		return
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	redirectURL, err := p.service.Purchase(ctxWithTimeout, ctx.MustGet("userID").(uuid.UUID), request)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, ginutil.Response{
		"The purchase procces has started",
		hypermedia.Links{
			{"redirect", hypermedia.LinkData{
				Href:   redirectURL,
				Method: "",
			}},
		},
	})
}

func (p *passengerHandler) purchaseFailed(ctx *gin.Context) {
	var sessionID = ctx.Param("id")
	if sessionID == "" {
		ginutil.HandlerProblemAbort(ctx, rfc7807.Unauthorized("unauthorized", "Unauthorized Error", "Unauthorized."))
		return
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	err := p.service.PurchaseFailed(ctxWithTimeout, sessionID)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.Redirect(http.StatusFound, config.FrontendURL()+"/connections")
}

func (p *passengerHandler) purchaseSucceded(ctx *gin.Context) {
	var sessionID = ctx.Param("id")
	if sessionID == "" {
		ginutil.HandlerProblemAbort(ctx, rfc7807.Unauthorized("unauthorized", "Unauthorized Error", "Unauthorized."))
		return
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	err := p.service.PurchaseSucceded(ctxWithTimeout, sessionID)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.Redirect(http.StatusFound, config.FrontendURL()+"/profile/tickets")
}
