package passenger

import (
	"maryan_api/internal/domain/passenger/service"

	"github.com/gin-gonic/gin"
)

type passengerHandler struct {
	service service.PassengerService
}

func (ch *passengerHandler) CreatePassenger(c *gin.Context) {}

func (ch *passengerHandler) GetPassenger(c *gin.Context) {}

func (ch *passengerHandler) GetPassengers(c *gin.Context) {}

func (ch *passengerHandler) UpdatePassenger(c *gin.Context) {}

func (ch *passengerHandler) DeletePassenger(c *gin.Context) {}

// ----------------Handlers Initialization Function---------------------

func newPassengerHandler(passenger service.PassengerService) passengerHandler {
	return passengerHandler{passenger}
}
