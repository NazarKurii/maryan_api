package passenger

import (
	"maryan_api/internal/domain/passenger/repo"
	"maryan_api/internal/domain/passenger/service"
	"maryan_api/pkg/auth"
	ginutil "maryan_api/pkg/ginutils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Admin struct {
	passenger passengerHandler
}

type Customer struct {
	passenger passengerHandler
}

func RegisterRoutes(db *gorm.DB, s *gin.Engine, client *http.Client) {

	customerRouter := ginutil.CreateAuthRouter("/customer", auth.Customer.SecretKey(), s)
	customer := Customer{
		newPassengerHandler(service.NewPassengerService(repo.NewPassengerRepoMysql(db), client)),
	}

	//--------------------PassengerRoutes---------------------------------

	customerRouter.POST("/passenger", customer.passenger.CreatePassenger)
	customerRouter.GET("/passenger/:id", customer.passenger.GetPassenger)
	customerRouter.GET("/passengers/:page/:size/:orderBy/:orderWay", customer.passenger.GetPassengers)
	customerRouter.PUT("/passenger", customer.passenger.UpdatePassenger)
	customerRouter.DELETE("/passenger/:id", customer.passenger.DeletePassenger)
}
