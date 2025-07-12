package http

import (
	"context"
	"maryan_api/config"
	"maryan_api/internal/domain/passenger/service"
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

type passengerHandler struct {
	service service.Passenger
}

func (p *passengerHandler) CreatePassenger(ctx *gin.Context) {
	var passenger entity.Passenger
	err := ctx.ShouldBindJSON(&passenger)
	if err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest(
			"passenger-parsing",
			"Passenger Parsing Error",
			err.Error(),
		))
		return
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	id, err := p.service.Create(ctxWithTimeout, passenger, ctx.MustGet("userID").(uuid.UUID))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, ginutil.Response{
		"The passenger has successfuly been created.",
		hypermedia.Links{
			hypermedia.Link{
				"self": hypermedia.Href{
					config.APIURL() + "/customer/passenger/" + id.String(),
					"GET",
				},
			},
			deletePassengerLink,
			updatePassengerLink,
			getPassengersLink,
		},
	})
}

func (p *passengerHandler) GetPassenger(ctx *gin.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	passenger, err := p.service.GetPassenger(ctxWithTimeout, ctx.Param("id"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, struct {
		Passenger entity.Passenger `json:"passenger"`
		ginutil.Response
	}{
		passenger,
		ginutil.Response{
			"The passenger has successfuly been created.",
			hypermedia.Links{
				deletePassengerLink,
				updatePassengerLink,
				getPassengersLink,
			},
		},
	})
}

func (p *passengerHandler) GetPassengers(ctx *gin.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()
	passengers, links, err := p.service.GetPassengers(ctxWithTimeout, dbutil.PaginationStr{
		"customer/passengers",
		ctx.DefaultQuery("page", "0"),
		ctx.DefaultQuery("size", "20"),
		ctx.DefaultQuery("order_by", ""),
		ctx.DefaultQuery("order_way", "ASC"),
		ctx.DefaultQuery("search", ""),
	}, ctx.MustGet("userID").(uuid.UUID))

	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		Links      hypermedia.Links   `json:"links"`
		Passengers []entity.Passenger `json:"passengers"`
	}{
		ginutil.Response{
			"The passengers have successfuly beeen found.",
			hypermedia.Links{
				deletePassengerLink,
				updatePassengerLink,
				getPassengerLink,
			},
		},
		links,
		passengers,
	})

}

func (p *passengerHandler) UpdatePassenger(ctx *gin.Context) {
	var passenger entity.Passenger
	err := ctx.ShouldBindJSON(&passenger)
	if err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest(
			"passenger-parsing",
			"Passenger Parsing Error",
			err.Error(),
		))
		return
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	id, err := p.service.Update(ctxWithTimeout, passenger)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, ginutil.Response{
		"The passenger has successfuly been updated.",
		hypermedia.Links{
			hypermedia.Link{
				"self": hypermedia.Href{
					config.APIURL() + "/passenger/" + id.String(),
					"GET",
				},
			},
			deletePassengerLink,
			updatePassengerLink,
			getPassengersLink,
		},
	})
}

func (p *passengerHandler) DeletePassenger(ctx *gin.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	err := p.service.Delete(ctxWithTimeout, ctx.Param("id"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated,
		ginutil.Response{
			"The passenger has successfuly been deleted.",
			hypermedia.Links{
				createPassengerLink,
				getPassengersLink,
			},
		},
	)
}

// ----------------Handlers Initialization Function---------------------

func newPassengerHandler(service service.Passenger) passengerHandler {
	return passengerHandler{service}
}
