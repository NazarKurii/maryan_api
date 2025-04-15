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

	"github.com/d3code/uuid"
	"github.com/gin-gonic/gin"
)

type addressHandler struct {
	service service.Address
}

// func (a *addressHandler) CreateAddress(ctx *gin.Context) {
// 	var address entity.NewAddress
// 	err := ctx.ShouldBindJSON(&address)
// 	if err != nil {
// 		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest(
// 			"address-parsing",
// 			"Address Parsing Error",
// 			err.Error(),
// 		))
// 		return
// 	}

// 	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
// 	defer cancel()

// 	id, err := a.service.Create(ctxWithTimeout, address, ctx.MustGet("userID").(uuid.UUID))
// 	if err != nil {
// 		ginutil.ServiceErrorAbort(ctx, err)
// 		return
// 	}

// 	ctx.JSON(http.StatusCreated, ginutil.Response{
// 		"The address has successfuly been created.",
// 		hypermedia.Links{

// 			hypermedia.Link{
// 				"self", hypermedia.LinkData{
// 					config.APIURL() + "/address/" + id.String(),
// 					"GET",
// 				},
// 			},

// 			deleteAddressLink,
// 			updateAddressLink,
// 			getAddressLink,
// 		},
// 	})
// }

func (a *addressHandler) GetAddress(ctx *gin.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	address, err := a.service.GetByID(ctxWithTimeout, ctx.Param("id"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, struct {
		Address entity.Address `json:"address"`
		ginutil.Response
	}{
		address,
		ginutil.Response{
			"The address has successfuly been created.",
			hypermedia.Links{
				deleteAddressLink,
				updateAddressLink,
				getAddressLink,
			},
		},
	})
}

func (a *addressHandler) GetAddresses(ctx *gin.Context) {

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()
	addresses, links, err := a.service.GetAddresses(ctxWithTimeout, dbutil.PaginationStr{
		"/customer/addresses",
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
		Links     hypermedia.Links `json:"links"`
		Addresses []entity.Address `json:"addresses"`
	}{
		ginutil.Response{
			"The addresses have successfuly beeen found.",
			hypermedia.Links{
				deleteAddressLink,
				updateAddressLink,
				getAddressLink,
			},
		},
		links,
		addresses,
	})

}

func (a *addressHandler) UpdateAddress(ctx *gin.Context) {
	var address entity.Address
	err := ctx.ShouldBindJSON(&address)
	if err != nil {
		ginutil.HandlerProblemAbort(ctx, rfc7807.BadRequest(
			"address-parsing",
			"Address Parsing Error",
			err.Error(),
		))
		return
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	id, err := a.service.Update(ctxWithTimeout, address)
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, ginutil.Response{
		"The address has successfuly been updated.",
		hypermedia.Links{

			hypermedia.Link{
				"self", hypermedia.LinkData{
					config.APIURL() + "/address/" + id.String(),
					"GET",
				},
			},
			deleteAddressLink,
			updateAddressLink,
			getAddressLink,
		},
	})
}

func (a *addressHandler) DeleteAddress(ctx *gin.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	err := a.service.Delete(ctxWithTimeout, ctx.Param("id"))
	if err != nil {
		ginutil.ServiceErrorAbort(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated,
		ginutil.Response{
			"The address has successfuly been deleted.",
			hypermedia.Links{
				deleteAddressLink,
				updateAddressLink,
				getAddressLink,
			},
		},
	)
}

// ----------------Handlers Initialization Function---------------------

func newAddressHandler(address service.Address) addressHandler {
	return addressHandler{address}
}
