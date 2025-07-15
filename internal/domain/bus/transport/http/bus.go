package http

import (
	"context"
	"encoding/json"
	"maryan_api/config"

	"maryan_api/internal/domain/bus/service"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	ginutil "maryan_api/pkg/ginutils"
	"maryan_api/pkg/hypermedia"

	rfc7807 "maryan_api/pkg/problem"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ----------------------Admin Handler---------------------------
type busHandler struct {
	service service.Bus
}

func (b *busHandler) getAvailableBuses(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	buses, urls, err := b.service.GetAvailable(ctxWithTimeout, dbutil.PaginationStr{
		"admin/buses/available",
		ctx.DefaultQuery("page", "0"),
		ctx.DefaultQuery("size", "20"),
		ctx.DefaultQuery("order_by", "id"),
		ctx.DefaultQuery("order_way", "asc"),
		ctx.DefaultQuery("search", ""),
	}, ctx.DefaultQuery("from", ""), ctx.DefaultQuery("to", ""))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		Buses []entity.Bus `json:"buses"`
	}{ginutil.Response{
		"Users have successfuly been retrieved.",
		urls,
	},
		buses})
}

const (
	leadDriverType = iota
	assistantDriverType
)

func (b *busHandler) changeDriver(driverType int) func(ctx *gin.Context) {
	var serviceFunc func(ctx context.Context, busIDStr string, driverIDStr string) error
	if driverType == leadDriverType {
		serviceFunc = b.service.ChangeDriver(service.LeadDriver)
	} else {
		serviceFunc = b.service.ChangeDriver(service.AssistantDriver)
	}

	return func(ctx *gin.Context) {
		var request struct {
			DriverID string `json:"driverId"`
		}

		err := ctx.ShouldBindJSON(&request)
		if err != nil {
			ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("body-parsing", "Body Parsing Error", err.Error()))
			return
		}

		ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
		defer cancel()

		err = serviceFunc(ctxWithTimeout, request.DriverID, ctx.Param("id"))
		if err != nil {
			ginutil.ServiceErrorAbort(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, ginutil.Response{
			"The driver has successfuly been changed",
			hypermedia.Links{},
		})
	}
}

func (b *busHandler) setBusSchedule(ctx *gin.Context) {
	var request struct {
		Schedule []entity.BusAvailability `json:"schedule"`
	}

	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest("body-parsing", "Body Parsing Error", err.Error()))
		return
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	err = b.service.SetSchedule(ctxWithTimeout, request.Schedule)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, ginutil.Response{
		"The schedule of the bus has successfuly been updated.",
		hypermedia.Links{},
	})
}

func (b *busHandler) createBus(ctx *gin.Context) {

	form, err := ctx.MultipartForm()
	if err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest(
			"form-parsing-error",
			"Form Parsing Error",
			err.Error(),
		))
		return
	}

	images, ok := form.File["images"]
	if !ok {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest(
			"form-parsing-error",
			"Form Parsing Error",
			"No images atached.",
		))
		return
	}

	jsonBus := ctx.PostForm("bus")
	var bus entity.NewBus
	if err := json.Unmarshal([]byte(jsonBus), &bus); err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest(
			"body-parsing-error",
			"Body Parsing Error",
			err.Error(),
		))
		return
	}

	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	id, err := b.service.Create(ctxWithTimeout, bus, images, ctx.SaveUploadedFile)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, ginutil.Response{
		"The bus has successfuly been created",
		hypermedia.Links{
			hypermedia.Link{
				"self", hypermedia.LinkData{
					config.APIURL() + "/admin/bus/" + id.String(),
					"GET",
				},
			},
			deleteBusLink,
		},
	})

}

func (b *busHandler) getBus(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	bus, err := b.service.GetByID(ctxWithTimeout, ctx.Param("id"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusFound, struct {
		ginutil.Response
		Bus entity.EmployeeBus `json:"bus"`
	}{
		ginutil.Response{
			"The bus has successfuly been found",
			hypermedia.Links{
				deleteBusLink,
			},
		},
		bus,
	})
}

func (b *busHandler) getBuses(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	buses, urls, err := b.service.GetBuses(ctxWithTimeout, dbutil.PaginationStr{
		"admin/buses",
		ctx.DefaultQuery("page", "1"),
		ctx.DefaultQuery("size", "10"),
		ctx.DefaultQuery("order_by", "created_at"),
		ctx.DefaultQuery("order_way", "ASC"),
		ctx.DefaultQuery("search", ""),
	})

	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		Buses []entity.Bus `json:"buses"`
	}{
		ginutil.Response{
			"The buses have successfuly been found",
			urls,
		},
		buses,
	})
}

func (b *busHandler) deleteBus(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	err := b.service.Delete(ctxWithTimeout, ctx.Param("id"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusFound, ginutil.Response{
		"The bus has successfuly been deleted",
		hypermedia.Links{
			createBusLink,
		},
	})
}

// ----------------Handlers Initialization Functions---------------------
func newBusHandler(bus service.Bus) busHandler {
	return busHandler{bus}
}
