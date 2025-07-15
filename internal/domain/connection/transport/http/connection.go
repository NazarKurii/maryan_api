package http

import (
	"context"
	"maryan_api/internal/domain/connection/service"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	ginutil "maryan_api/pkg/ginutils"
	"maryan_api/pkg/hypermedia"
	rfc7807 "maryan_api/pkg/problem"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type adminHandler struct {
	service service.AdminConnection
}

func (ch *adminHandler) GetByID(ctx *gin.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	connection, err := ch.service.GetByID(ctxWithTimeout, ctx.Param("id"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		Connection entity.Connection `json:"connection"`
		ginutil.Response
	}{
		connection,
		ginutil.Response{
			Message: "The connection has successfuly been found.",
		},
	})
}

func (ch *adminHandler) GetConnections(ctx *gin.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	connections, urls, err := ch.service.GetConnections(
		ctxWithTimeout,
		dbutil.PaginationStr{
			Path:   "/admin/connections",
			Page:   ctx.DefaultQuery("page", "0"),
			Size:   ctx.DefaultQuery("size", "10"),
			Search: ctx.DefaultQuery("search", ""),
		},
		ctx.DefaultQuery("complete", "false"),
	)

	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		Connections []entity.ConnectionSimplified `json:"connections"`
		Urls        hypermedia.Links              `json:"urls"`
		ginutil.Response
	}{
		connections,
		urls,
		ginutil.Response{
			Message: "The connections has successfuly been found.",
		},
	})
}

func (ch *adminHandler) RegisterUpdate(ctx *gin.Context) {
	var update entity.ConnectionUpdate

	if err := ctx.ShouldBindJSON(&update); err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("connection-update-data", "Invalid Connection Update Datat Error", err.Error()))
		return
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	err := ch.service.RegisterUpdate(ctxWithTimeout, update)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, ginutil.Response{
		Message: "The connection update has successfuly been registered.",
	})

}

func newAdminHandler(service service.AdminConnection) adminHandler {
	return adminHandler{service}
}

type customerHandler struct {
	service service.CustomerConnection
}

func (ch *customerHandler) GetByID(ctx *gin.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	connection, err := ch.service.GetByID(ctxWithTimeout, ctx.Param("id"))
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

func (ch *customerHandler) GetConnections(ctx *gin.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	connections, urls, err := ch.service.GetConnections(
		ctxWithTimeout,
		ctx.MustGet("userID").(uuid.UUID),
		dbutil.PaginationStr{
			Path:   "/customer/connections",
			Page:   ctx.DefaultQuery("page", "0"),
			Size:   ctx.DefaultQuery("size", "10"),
			Search: ctx.DefaultQuery("search", ""),
		},
		ctx.DefaultQuery("complete", "false"),
	)

	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		Connections []entity.CustomerConnection `json:"connections"`
		Urls        hypermedia.Links            `json:"urls"`
		ginutil.Response
	}{
		connections,
		urls,
		ginutil.Response{
			Message: "The connections has successfuly been found.",
		},
	})
}

func (ch *customerHandler) FindConnections(ctx *gin.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	response, err := ch.service.FindConnections(
		ctxWithTimeout,
		entity.FindConnectionsRequestJSON{
			From:      ctx.Param("from"),
			To:        ctx.Param("to"),
			Date:      ctx.Param("date"),
			Adults:    ctx.Param("adults"),
			Children:  ctx.Param("children"),
			Teenagers: ctx.Param("teenagers"),
			Range:     ctx.DefaultQuery("range", "5"),
		},
	)

	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		entity.FindConnectionsResponse
		ginutil.Response
	}{
		response,
		ginutil.Response{
			Message: "The connections has successfuly been found.",
		},
	})
}

func newCustomerHandler(service service.CustomerConnection) customerHandler {
	return customerHandler{service}
}
