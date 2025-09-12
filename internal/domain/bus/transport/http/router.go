package http

import (
	"maryan_api/internal/domain/bus/repo"
	"maryan_api/internal/domain/bus/service"
	"maryan_api/pkg/auth"
	ginutil "maryan_api/pkg/ginutils"
	"maryan_api/pkg/hypermedia"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(db *gorm.DB, s *gin.Engine, client *http.Client) {
	adminRouter := ginutil.CreateAuthRouter("/admin", auth.Admin.SecretKey(), s)
	handler := newBusHandler(service.NewBusService(repo.NewBusRepo(db), repo.NewDriverRepo(db)))

	//-----------------------Bus Routes------------------------------------
	adminRouter.POST("/bus", handler.createBus)
	adminRouter.GET("/bus/:id", handler.getBus)
	adminRouter.GET("/buses", handler.getBuses)
	adminRouter.DELETE("/bus", handler.deleteBus)
	adminRouter.POST("/bus/schedule", handler.setBusSchedule)
	adminRouter.PATCH("/bus/:id/lead-driver", handler.changeDriver(leadDriverType))
	adminRouter.PATCH("/bus/:id/assistant-driver", handler.changeDriver(assistantDriverType))
	adminRouter.GET("/buses/available", handler.getAvailableBuses)
}

// -------------Links-----------------
var (
	createBusLink = hypermedia.Link{
		Name: "createBus",
		Data: hypermedia.LinkData{Href: "/bus", Method: "POST"},
	}

	listBusesLink = hypermedia.Link{
		Name: "listBuses",
		Data: hypermedia.LinkData{Href: "/bus", Method: "GET"},
	}

	deleteBusLink = hypermedia.Link{
		Name: "deleteBus",
		Data: hypermedia.LinkData{Href: "/bus", Method: "DELETE"},
	}
)
