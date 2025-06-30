package trip

import (
	"encoding/json"
	"maryan_api/config"
	"maryan_api/pkg/auth"
	ginutil "maryan_api/pkg/ginutils"
	"maryan_api/pkg/hypermedia"
	rfc7807 "maryan_api/pkg/problem"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(db *gorm.DB, s *gin.Engine, client *http.Client) {

	adminRouter := ginutil.CreateAuthRouter("/admin", auth.Admin.SecretKey(), s)
	admin := newAdminHandler(newBusService(newBusRepoMySql(db)))

	customerRouter := ginutil.CreateAuthRouter("/customer", auth.Customer.SecretKey(), s)
	customer := newCustomerHandler(newPassengerService(newPassengerRepoMysql(db), client))

	//-----------------------BusRoutes------------------------------------
	adminRouter.POST("/bus", admin.createBus)
	adminRouter.GET("/bus/:id", admin.getBus)
	adminRouter.GET("/buses/:page/:size/:simplify", admin.getBuses)
	adminRouter.DELETE("/bus", admin.deleteBus)
	adminRouter.PATCH("/bus/make-inactive/:id", admin.makeBusInactive)
	adminRouter.PATCH("/bus/make-active/:id", admin.makeBusActive)

	//--------------------PassengerRoutes---------------------------------

}

type adminHandler struct {
	bus busService
}

func (ah *adminHandler) createBus(c *gin.Context) {

	form, err := c.MultipartForm()
	if err != nil {
		ginutil.HandlerProblemAbort(c, rfc7807.BadRequest(
			"form-parsing-error",
			"Form Parsing Error",
			err.Error(),
		))
		return
	}

	images, ok := form.File["images"]
	if !ok {
		ginutil.HandlerProblemAbort(c, rfc7807.BadRequest(
			"form-parsing-error",
			"Form Parsing Error",
			"No images atached.",
		))
		return
	}

	jsonBus := c.PostForm("bus")
	var bus Bus
	if err := json.Unmarshal([]byte(jsonBus), &bus); err != nil {
		ginutil.HandlerProblemAbort(c, rfc7807.BadRequest(
			"body-parsing-error",
			"Body Parsing Error",
			err.Error(),
		))
		return
	}

	id, err := ah.bus.create(bus, images)
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusCreated, ginutil.Response{
		"The bus has successfuly been created",
		hypermedia.Links{
			hypermedia.Link{"self": hypermedia.Href{config.APIURL() + "/admin/bus/" + id.String(), http.MethodGet}},
			deleteBusLink,
			makeBusActiveLink,
			makeBusInactiveLink,
		},
	})

}

func (ah *adminHandler) getBus(c *gin.Context) {
	bus, err := ah.bus.get(c.Param("id"))
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusFound, struct {
		ginutil.Response
		Bus Bus `json:"bus"`
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

func (ah *adminHandler) getBuses(c *gin.Context) {
	buses, urls, err := ah.bus.getBuses(c.Param("page"), c.Param("size"))
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusFound, struct {
		ginutil.Response
		Buses []Bus            `json:"buses"`
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

func (ah *adminHandler) deleteBus(c *gin.Context) {
	err := ah.bus.delete(c.Param("id"))
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusFound, ginutil.Response{
		"The bus has successfuly been deleted",
		hypermedia.Links{
			createBusLink,

			makeBusActiveLink,
			makeBusInactiveLink,
		},
	})
}

func (ah *adminHandler) makeBusActive(c *gin.Context) {
	err := ah.bus.makeActive(c.Param("id"))
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusFound, ginutil.Response{
		"The bus has successfuly been made active",
		hypermedia.Links{
			deleteBusLink,
			makeBusInactiveLink,
		},
	})
}

func (ah *adminHandler) makeBusInactive(c *gin.Context) {
	err := ah.bus.makeInActive(c.Param("id"))
	if err != nil {
		ginutil.ServiceErrorAbort(c, err)
		return
	}

	c.JSON(http.StatusFound, ginutil.Response{
		"The bus has successfuly been made inactive",
		hypermedia.Links{
			deleteBusLink,
			makeBusActiveLink,
		},
	})
}

type CustomerHandler struct {
	service passengerService
}

// ----------------Handlers Initialization Functions---------------------
func newAdminHandler(bus busService) adminHandler {
	return adminHandler{bus}
}

func newCustomerHandler(passenger passengerService) CustomerHandler {
	return CustomerHandler{passenger}
}
