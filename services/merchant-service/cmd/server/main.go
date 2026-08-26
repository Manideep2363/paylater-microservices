package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"paylater/services/merchant-service/internal/database"
	"paylater/services/merchant-service/internal/handler"
	"paylater/services/merchant-service/internal/repository"
	"paylater/services/merchant-service/internal/routes"
	"paylater/services/merchant-service/internal/service"
	"paylater/shared/config"
)

func main() {
	cfg := config.LoadConfig()
	applyMerchantServiceDefaults(cfg)

	dbConn, err := database.NewMySQL(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer dbConn.Close()

	repo := repository.NewSQLRepository(dbConn)
	merchantService := service.NewMerchantService(repo)
	merchantHandler := handler.NewMerchantHandler(merchantService)

	router := gin.Default()
	routes.Setup(router, merchantHandler, cfg.JWTSecret)

	log.Printf("merchant-service listening on :%s (db=%s)", cfg.ServerPort, cfg.DBName)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}

func applyMerchantServiceDefaults(cfg *config.Config) {
	if os.Getenv("DB_NAME") == "" {
		cfg.DBName = "paylater_merchants"
	}
	if os.Getenv("SERVER_PORT") == "" {
		cfg.ServerPort = "8083"
	}
}
