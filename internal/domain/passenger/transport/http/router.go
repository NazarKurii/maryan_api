package http

import (
	"maryan_api/internal/domain/passenger/repo"
	"maryan_api/internal/domain/passenger/service"
	"maryan_api/pkg/auth"
	ginutil "maryan_api/pkg/ginutils"
	"maryan_api/pkg/hypermedia"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(db *gorm.DB, s *gin.Engine, client *http.Client) {
	handler := newPassengerHandler(service.NewPassengerService(repo.NewPassengerRepoMysql(db), client))

	customerRouter := ginutil.CreateAuthRouter("/customer", auth.Customer.SecretKey(), s)

	customerRouter.POST("/passenger", handler.CreatePassenger)
	customerRouter.GET("/passenger/:id", handler.GetPassenger)
	customerRouter.GET("/passengers", handler.GetPassengers)
	customerRouter.PUT("/passenger", handler.UpdatePassenger)
	customerRouter.DELETE("/passenger/:id", handler.DeletePassenger)
}

var (
	createPassengerLink = hypermedia.Link{
		Name: "createPassenger",
		Data: hypermedia.LinkData{Href: "/passenger", Method: "POST"},
	}

	getPassengerLink = hypermedia.Link{
		Name: "getPassenger",
		Data: hypermedia.LinkData{Href: "/passenger/:id", Method: "GET"},
	}

	getPassengersLink = hypermedia.Link{
		Name: "getPassengers",
		Data: hypermedia.LinkData{Href: "/passengers/:page/:size/:orderBy/:orderWay", Method: "GET"},
	}

	listPassengersLink = hypermedia.Link{
		Name: "listPassengers",
		Data: hypermedia.LinkData{Href: "/passengers/:page/:size/:orderBy/:orderWay", Method: "GET"},
	}

	updatePassengerLink = hypermedia.Link{
		Name: "updatePassenger",
		Data: hypermedia.LinkData{Href: "/passenger", Method: "PATCH"},
	}

	deletePassengerLink = hypermedia.Link{
		Name: "deletePassenger",
		Data: hypermedia.LinkData{Href: "/passenger/:id", Method: "DELETE"},
	}
)
