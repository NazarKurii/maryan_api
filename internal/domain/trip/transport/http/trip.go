package http

import (
	"maryan_api/config"
	"maryan_api/internal/domain/trip/service"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	ginutil "maryan_api/pkg/ginutils"
	"maryan_api/pkg/hypermedia"
	rfc7807 "maryan_api/pkg/problem"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type tripHandler struct {
	service service.Trip
}

func (h tripHandler) Create(ctx *gin.Context) {
	var trip entity.Trip

	err := ctx.ShouldBindJSON(&trip)
	if err != nil {
		ginutil.HandlerProblemAbort(
			ctx,
			rfc7807.BadRequest(
				"trip-data",
				"Trip Data Error",
				err.Error()),
		)
		return
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	id, err := h.service.Create(ctxWithTimeout, trip)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, ginutil.Response{
		"The trip has successfuly been created.",
		hypermedia.Links{
			hypermedia.Link{

				"self", hypermedia.LinkData{config.APIURL() + "/admin/trip/" + id.String(), "GET"},
			},
		},
	})
}

func (h tripHandler) CreateTest(ctx *gin.Context) {

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*1000)
	defer cancel()

	err := h.service.CreateTestTrips(ctxWithTimeout)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, ginutil.Response{
		"The trips has successfuly been created.",
		hypermedia.Links{},
	})
}

func (h tripHandler) GetByID(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	trip, err := h.service.GetByID(ctxWithTimeout, ctx.Param("id"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		Trip entity.Trip `json:"trip"`
		ginutil.Response
	}{
		trip,
		ginutil.Response{
			Message: "The trip ha successfuly been found.",
		},
	})
}

func (h tripHandler) GetTrips(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	trips, urls, err := h.service.GetTrips(ctxWithTimeout, dbutil.PaginationStr{
		Path: "admin/trips",
		Page: ctx.DefaultQuery("page", "0"),
		Size: ctx.DefaultQuery("page", "5"),
	})

	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		Trips []entity.TripSimplified `json:"trips"`
		Urls  hypermedia.Links        `json:"urls"`
		ginutil.Response
	}{
		trips,
		urls,
		ginutil.Response{
			Message: "The trip ha successfuly been found.",
		},
	})
}

func (h tripHandler) RegisterUpdate(ctx *gin.Context) {
	var update entity.TripUpdate

	err := ctx.ShouldBindJSON(&update)
	if err != nil {
		ginutil.HandlerProblemAbort(
			ctx,
			rfc7807.BadRequest(
				"trip-update-data",
				"Trip Update Data Error",
				err.Error()),
		)
		return
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	err = h.service.RegisterUpdate(ctxWithTimeout, update)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, ginutil.Response{
		Message: "The update has successfuly been registered.",
	})

}

func newTripHandler(service service.Trip) tripHandler {
	return tripHandler{service}
}
