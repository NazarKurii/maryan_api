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
	handler := newAddressHandler(service.NewAddressService(repo.NewAddressRepo(db), client))

	customerRouter := ginutil.CreateAuthRouter("/customer", auth.Customer.SecretKey(), s)

	//--------------------PassengerRoutes---------------------------------

	// customerRouter.POST("/Address", handler.CreateAddress)
	customerRouter.GET("/Address/:id", handler.GetAddress)
	customerRouter.GET("/Addresses", handler.GetAddresses)
	customerRouter.PUT("/Address", handler.UpdateAddress)
	customerRouter.DELETE("/Address/:id", handler.DeleteAddress)
}

var (
	createAddressLink = hypermedia.Link{
		Name: "createAddress",
		Data: hypermedia.LinkData{Href: "/customer/address", Method: "POST"},
	}

	getAddressLink = hypermedia.Link{
		Name: "getAddress",
		Data: hypermedia.LinkData{Href: "/customer/address/:id", Method: "GET"},
	}

	getAddressesLink = hypermedia.Link{
		Name: "getAddresses",
		Data: hypermedia.LinkData{Href: "/customer/addresses/:page/:size/:orderBy/:orderWay", Method: "GET"},
	}

	listAddressesLink = hypermedia.Link{
		Name: "listAddresses",
		Data: hypermedia.LinkData{Href: "/customer/addresses/:page/:size/:orderBy/:orderWay", Method: "GET"},
	}

	updateAddressLink = hypermedia.Link{
		Name: "updateAddress",
		Data: hypermedia.LinkData{Href: "/customer/address", Method: "PATCH"},
	}

	deleteAddressLink = hypermedia.Link{
		Name: "deleteAddress",
		Data: hypermedia.LinkData{Href: "/customer/address/:id", Method: "DELETE"},
	}
)
