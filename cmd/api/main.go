package main

import (
	"os"

	"github.com/nazarkurii/marshrutka_api/config"
	"github.com/nazarkurii/marshrutka_api/internal/infrastructure/clients/stripe"
	dataStore "github.com/nazarkurii/marshrutka_api/internal/infrastructure/persistence"
	"github.com/nazarkurii/marshrutka_api/internal/infrastructure/router"
	"github.com/nazarkurii/marshrutka_api/pkg/languages"
	"github.com/nazarkurii/marshrutka_api/pkg/timezone"

	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadConfig("/home/nazar/nazar/marshrutka/api/.env")
	timezone.Load()

	db := dataStore.Init()
	// dataStore.Migrate(db)
	config.LoadCountries(db)

	stripe.InitStripe()
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowOrigins: []string{"https://marshrutka.store", "http://marshrutka.store", "http://localhost:3000"},
		AllowHeaders: []string{"Authorization", "Content-Type", "X-Email-Access-Token", "X-Customer-Update-Token", "Content-Language"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))

	server.Use(languages.GinMiddlewear)
	client := http.DefaultClient
	router.RegisterRoutes(server, db, client)
	server.Static("/imgs", "../../static/images")
	server.GET("/health", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	err := server.Run(os.Getenv("PORT"))
	if err != nil {
		panic(err)
		//
	}
}
