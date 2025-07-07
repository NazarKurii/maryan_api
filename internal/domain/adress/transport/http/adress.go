package http

import (
	"context"
	"maryan_api/config"
	"maryan_api/internal/domain/adress/service"
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

type adressHandler struct {
	service service.Adress
}

func (a *adressHandler) CreateAdress(ctx *gin.Context) {
	var adress entity.Adress
	err := ctx.ShouldBindJSON(&adress)
	if err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest(
			"adress-parsing",
			"Adress Parsing Error",
			err.Error(),
		))
		return
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	id, err := a.service.Create(ctxWithTimeout, adress, ctx.MustGet("userID").(uuid.UUID))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, ginutil.Response{
		"The adress has successfuly been created.",
		hypermedia.Links{
			hypermedia.Link{
				"self": hypermedia.Href{
					config.APIURL() + "/customer/adress/" + id.String(),
					"GET",
				},
			},
			deleteAddressLink,
			updateAddressLink,
			getAddressLink,
		},
	})
}

func (a *adressHandler) GetAdress(ctx *gin.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	adress, err := a.service.GetByID(ctxWithTimeout, ctx.Param("id"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, struct {
		Adress entity.Adress `json:"adress"`
		ginutil.Response
	}{
		adress,
		ginutil.Response{
			"The adress has successfuly been created.",
			hypermedia.Links{
				deleteAddressLink,
				updateAddressLink,
				getAddressLink,
			},
		},
	})
}

func (a *adressHandler) GetAdresses(ctx *gin.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()
	adresses, links, err := a.service.GetAdresses(ctxWithTimeout, dbutil.PaginationStr{
		"/customer/adresses",
		ctx.Param("page"),
		ctx.Param("size"),
		ctx.Param("order_by"),
		ctx.Param("order_way"),
	}, ctx.MustGet("userID").(uuid.UUID))

	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ginutil.Response
		Links    hypermedia.Links `json:"links"`
		Adresses []entity.Adress  `json:"adresses"`
	}{
		ginutil.Response{
			"The adresses have successfuly beeen found.",
			hypermedia.Links{
				deleteAddressLink,
				updateAddressLink,
				getAddressLink,
			},
		},
		links,
		adresses,
	})

}

func (a *adressHandler) UpdateAdress(ctx *gin.Context) {
	var adress entity.Adress
	err := ctx.ShouldBindJSON(&adress)
	if err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest(
			"adress-parsing",
			"Adress Parsing Error",
			err.Error(),
		))
		return
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	id, err := a.service.Update(ctxWithTimeout, adress)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, ginutil.Response{
		"The adress has successfuly been updated.",
		hypermedia.Links{
			hypermedia.Link{
				"self": hypermedia.Href{
					config.APIURL() + "/adress/" + id.String(),
					"GET",
				},
			},
			deleteAddressLink,
			updateAddressLink,
			getAddressLink,
		},
	})
}

func (a *adressHandler) DeleteAdress(ctx *gin.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	err := a.service.Delete(ctxWithTimeout, ctx.Param("id"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated,
		ginutil.Response{
			"The adress has successfuly been deleted.",
			hypermedia.Links{
				createAddressLink,
				getAddressLink,
			},
		},
	)
}

// ----------------Handlers Initialization Function---------------------

func newAdressHandler(adress service.Adress) adressHandler {
	return adressHandler{adress}
}
