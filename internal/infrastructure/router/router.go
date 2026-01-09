package router

import (
	"net/http"

	adress "github.com/nazarkurii/marshrutka_api/internal/domain/adress/transport/http"
	bus "github.com/nazarkurii/marshrutka_api/internal/domain/bus/transport/http"
	connection "github.com/nazarkurii/marshrutka_api/internal/domain/connection/transport/http"
	"github.com/nazarkurii/marshrutka_api/internal/domain/documents"
	parcel "github.com/nazarkurii/marshrutka_api/internal/domain/parcel/transport/http"
	passenger "github.com/nazarkurii/marshrutka_api/internal/domain/passenger/transport/http"
	ticket "github.com/nazarkurii/marshrutka_api/internal/domain/tickets/transport/http"
	trip "github.com/nazarkurii/marshrutka_api/internal/domain/trip/transport/http"
	user "github.com/nazarkurii/marshrutka_api/internal/domain/user/transport/http"
	ginutil "github.com/nazarkurii/marshrutka_api/pkg/ginutils"

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
	documents.RegisterRoutes(db, s, client)
	parcel.RegisterRoutes(db, s, client)
}
