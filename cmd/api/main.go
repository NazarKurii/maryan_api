package main

import (
	"maryan_api/config"
	"maryan_api/internal/infrastructure/clients/stripe"
	dataStore "maryan_api/internal/infrastructure/persistence"
	"maryan_api/internal/infrastructure/router"
	"maryan_api/pkg/timezone"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadConfig("../../.env")
	timezone.Load()

	db := dataStore.Init()
	dataStore.Migrate(db)
	config.LoadCountries(db)

	stripe.InitStripe()
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Authorization", "Content-Type", "X-Email-Access-Token", "X-Customer-Update-Token"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))
	client := http.DefaultClient
	router.RegisterRoutes(server, db, client)
	server.Static("/imgs", "../../static/imgs")

	server.Run(":8080")
}
