package router

import (
	adress "maryan_api/internal/domain/adress/transport/http"
	bus "maryan_api/internal/domain/bus/transport/http"
	connection "maryan_api/internal/domain/connection/transport/http"
	passenger "maryan_api/internal/domain/passenger/transport/http"
	ticket "maryan_api/internal/domain/tickets/transport/http"
	trip "maryan_api/internal/domain/trip/transport/http"
	user "maryan_api/internal/domain/user/transport/http"
	ginutil "maryan_api/pkg/ginutils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(s *gin.Engine, db *gorm.DB, client *http.Client) {
	s.Use(ginutil.LogMiddlewear(db))

	passenger.RegisterRoutes(db, s, client)
	user.RegisterRoutes(db, s, client)
	bus.RegisterRoutes(db, s, client)
	adress.RegisterRoutes(db, s, client)
	connection.RegisterRoutes(db, s, client)
	trip.RegisterRoutes(db, s, client)
	ticket.RegisterRoutes(db, s, client)
}
