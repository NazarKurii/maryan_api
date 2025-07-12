package http

import (
	"maryan_api/internal/domain/adress/repo"
	"maryan_api/internal/domain/adress/service"
	"maryan_api/pkg/auth"
	ginutil "maryan_api/pkg/ginutils"
	"maryan_api/pkg/hypermedia"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(db *gorm.DB, s *gin.Engine, client *http.Client) {
	handler := newAdressHandler(service.NewAdressService(repo.NewPassengerRepo(db), client))

	customerRouter := ginutil.CreateAuthRouter("/customer", auth.Customer.SecretKey(), s)

	//--------------------PassengerRoutes---------------------------------

	customerRouter.POST("/passenger", handler.CreateAdress)
	customerRouter.GET("/passenger/:id", handler.GetAdress)
	customerRouter.GET("/passengers", handler.GetAdress)
	customerRouter.PUT("/passenger", handler.UpdateAdress)
	customerRouter.DELETE("/passenger/:id", handler.DeleteAdress)
}

var createAddressLink = hypermedia.Link{
	"createAddress": {Href: "/customer/address", Method: "POST"},
}

var getAddressLink = hypermedia.Link{
	"getAddress": {Href: "/customer/address/:id", Method: "GET"},
}

var getAddressesLink = hypermedia.Link{
	"getAddress": {Href: "/customer/addresses/:page/:size/:orderBy/:orderWay", Method: "GET"},
}

var listAddressesLink = hypermedia.Link{
	"listAddresses": {Href: "/customer/addresses/:page/:size/:orderBy/:orderWay", Method: "GET"},
}

var updateAddressLink = hypermedia.Link{
	"updateAddress": {Href: "/customer/address", Method: "PATCH"},
}

var deleteAddressLink = hypermedia.Link{
	"deleteAddress": {Href: "/customer/address/:id", Method: "DELETE"},
}
