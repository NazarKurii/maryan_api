package main

import (
	"maryan_api/config"
	dataStore "maryan_api/internal/infrastructure/persistence"
	"maryan_api/internal/infrastructure/router"
	"maryan_api/pkg/timezone"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadConfig("../../.env")
	timezone.Load()

	db := dataStore.Init()
	dataStore.Migrate(db)

	server := gin.Default()
	client := http.DefaultClient
	router.RegisterRoutes(server, db, client)

	server.Run(":8080")
}
