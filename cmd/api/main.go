package main

import (
	"maryan_api/config"
	"maryan_api/internal/infrastructure/db"
	"maryan_api/internal/infrastructure/db/migrations"
	"maryan_api/internal/router"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadConfig("../../.env")
	db := db.Init()
	migrations.Migrate(db)

	server := gin.Default()
	client := http.DefaultClient
	router.RegisterRoutes(server, client, db)

	server.Run(":8080")
}
