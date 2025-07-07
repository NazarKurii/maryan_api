package http

import (
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
	var bus entity.Bus
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

	id, err := b.service.Create(ctxWithTimeout, bus, images)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, ginutil.Response{
		"The bus has successfuly been created",
		hypermedia.Links{
			hypermedia.Link{"self": hypermedia.Href{config.APIURL() + "/admin/bus/" + id.String(), http.MethodGet}},
			deleteBusLink,
			makeBusActiveLink,
			makeBusInactiveLink,
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
		Bus entity.Bus `json:"bus"`
	}{
		ginutil.Response{
			"The bus has successfuly been found",
			hypermedia.Links{
				deleteBusLink,
				makeBusActiveLink,
				makeBusInactiveLink,
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
		ctx.Param("page"),
		ctx.Param("size"),
		ctx.Param("order_by"),
		ctx.Param("order_way"),
	})

	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusFound, struct {
		ginutil.Response
		Buses []entity.Bus     `json:"buses"`
		Urls  hypermedia.Links `json:"urls"`
	}{
		ginutil.Response{
			"The buses have successfuly been found",
			hypermedia.Links{
				deleteBusLink,
				makeBusActiveLink,
				makeBusInactiveLink,
			},
		},
		buses,
		urls,
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

			makeBusActiveLink,
			makeBusInactiveLink,
		},
	})
}

func (b *busHandler) makeBusActive(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	err := b.service.MakeActive(ctxWithTimeout, ctx.Param("id"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusFound, ginutil.Response{
		"The bus has successfuly been made active",
		hypermedia.Links{
			deleteBusLink,
			makeBusInactiveLink,
		},
	})
}

func (b *busHandler) makeBusInactive(ctx *gin.Context) {
	ctxWithTimeout, cancel := ginutil.ContextWithTimeout(ctx, time.Second*20)
	defer cancel()

	err := b.service.MakeInActive(ctxWithTimeout, ctx.Param("id"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusFound, ginutil.Response{
		"The bus has successfuly been made inactive",
		hypermedia.Links{
			deleteBusLink,
			makeBusActiveLink,
		},
	})
}

// ----------------Handlers Initialization Functions---------------------
func newBusHandler(bus service.Bus) busHandler {
	return busHandler{bus}
}
