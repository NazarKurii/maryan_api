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
	customerRouter.GET("/passengers/:page/:size/:order_by/:order_way", handler.GetPassengers)
	customerRouter.PUT("/passenger", handler.UpdatePassenger)
	customerRouter.DELETE("/passenger/:id", handler.DeletePassenger)
}

var createPassengerLink = hypermedia.Link{
	"createPassenger": {Href: "/passenger", Method: "POST"},
}

var getPassengerLink = hypermedia.Link{
	"getPassenger": {Href: "/passenger/:id", Method: "GET"},
}

var getPassengersLink = hypermedia.Link{
	"getPassenger": {Href: "/passengers/:page/:size/:orderBy/:orderWay", Method: "GET"},
}

var listPassengersLink = hypermedia.Link{
	"listPassengers": {Href: "/passengers/:page/:size/:orderBy/:orderWay", Method: "GET"},
}

var updatePassengerLink = hypermedia.Link{
	"updatePassenger": {Href: "/passenger", Method: "PATCH"},
}

var deletePassengerLink = hypermedia.Link{
	"deletePassenger": {Href: "/passenger/:id", Method: "DELETE"},
}
