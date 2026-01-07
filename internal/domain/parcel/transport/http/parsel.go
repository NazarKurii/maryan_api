package http

import (
	"context"
	"net/http"
	"time"

	"github.com/nazarkurii/marshrutka_api/config"
	"github.com/nazarkurii/marshrutka_api/internal/domain/parcel/service"
	"github.com/nazarkurii/marshrutka_api/internal/entity"
	"github.com/nazarkurii/marshrutka_api/pkg/dbutil"
	ginutil "github.com/nazarkurii/marshrutka_api/pkg/ginutils"
	"github.com/nazarkurii/marshrutka_api/pkg/hypermedia"
	rfc7807 "github.com/nazarkurii/marshrutka_api/pkg/problem"

	"github.com/d3code/uuid"
	"github.com/gin-gonic/gin"
)

type parcelHandler struct {
	service service.Parcel
}

func newHandler(service service.Parcel) *parcelHandler {
	return &parcelHandler{service}
}

func (ch *parcelHandler) findConnections(ctx *gin.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	req := entity.FindParcelConnectionsRequest{
		From:   ctx.Param("from"),
		To:     ctx.Param("to"),
		Year:   ctx.Param("year"),
		Month:  ctx.Param("month"),
		Length: ctx.Param("length"),
		Height: ctx.Param("height"),
		Width:  ctx.Param("width"),
	}
	connections, err := ch.service.FindConnections(
		ctxWithTimeout,
		req,
	)

	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		Connections []entity.ConnectionParcel `json:"connections"`
		ginutil.Response
	}{
		connections,
		ginutil.Response{
			Message: "The connections has successfuly been found.",
		},
	})
}

func (p *parcelHandler) purchase(ctx *gin.Context) {
	var request entity.PurchaseParcelRequest

	err := ctx.ShouldBindJSON(&request)
	if err != nil {

		ginutil.HandlerProblemAbort(ctx, rfc7807.JSON(err.Error()))
		return
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	redirectURL, err := p.service.Purchase(ctxWithTimeout, ctx.MustGet("userID").(uuid.UUID), ctx.Param("id"), request)
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

func (p *parcelHandler) purchaseSucceded(ctx *gin.Context) {
	var sessionID = ctx.Param("id")
	if sessionID == "" {
		ginutil.HandlerProblemAbort(ctx, rfc7807.Unauthorized("unauthorized", "Unauthorized Error", "Unauthorized."))
		return
	}

	var token = ctx.Param("token")
	if token == "" {
		ginutil.HandlerProblemAbort(ctx, rfc7807.Unauthorized("unauthorized", "Unauthorized Error", "Unauthorized."))
		return
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	err := p.service.PurchaseSucceded(ctxWithTimeout, sessionID, token)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.Redirect(http.StatusFound, config.FrontendURL()+"/profile/parcels")
}

func (p *parcelHandler) getParcels(ctx *gin.Context) {

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	parcels, links, err := p.service.GetParcels(ctxWithTimeout, dbutil.PaginationStr{
		"/customer/parcels",
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
		Parcels []entity.CustomerParcel `json:"parcels"`
	}{
		ginutil.Response{
			"The parcels have successfuly beeen found.",
			hypermedia.Links{},
		},
		links,
		parcels,
	})

}

func (p *parcelHandler) purchaseFailed(ctx *gin.Context) {
	var sessionID = ctx.Param("id")
	if sessionID == "" {
		ginutil.HandlerProblemAbort(ctx, rfc7807.Unauthorized("unauthorized", "Unauthorized Error", "Unauthorized."))
		return
	}

	var token = ctx.Param("token")
	if token == "" {
		ginutil.HandlerProblemAbort(ctx, rfc7807.Unauthorized("unauthorized", "Unauthorized Error", "Unauthorized."))
		return
	}

	err := p.service.PurchaseFailed(ctx, sessionID, token)
	if err != nil {
		ctx.Redirect(http.StatusFound, config.FrontendURL()+"/internal-server-error")
		return
	}

	ctx.Redirect(http.StatusFound, config.FrontendURL()+"/parcel/failed/failed/failed/failed/failed/failed/true")
}

func (ch *parcelHandler) GetByID(ctx *gin.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()
	connection, err := ch.service.GetConnectionByID(ctxWithTimeout, ctx.Param("id"), ctx.Param("width"), ctx.Param("height"), ctx.Param("length"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		Connection entity.CustomerConnection `json:"connection"`
		ginutil.Response
	}{
		connection,
		ginutil.Response{
			Message: "The connection has successfuly been found.",
		},
	})
}
