package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"paylater/services/ledger-service/internal/client"
	"paylater/services/ledger-service/internal/config"
	"paylater/services/ledger-service/internal/database"
	"paylater/services/ledger-service/internal/handler"
	"paylater/services/ledger-service/internal/repository"
	"paylater/services/ledger-service/internal/routes"
	"paylater/services/ledger-service/internal/service"
)

func main() {
	cfg := config.Load()

	if cfg.InternalAPIToken == "" {
		log.Fatal("INTERNAL_API_TOKEN is required")
	}

	dbConn, err := database.NewMySQL(cfg.Config)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer dbConn.Close()

	store := repository.NewSQLStore(dbConn)
	userClient := client.NewUserClient(cfg.UserServiceURL, cfg.InternalAPIToken)
	merchantClient := client.NewMerchantClient(cfg.MerchantServiceURL, cfg.InternalAPIToken)

	ledgerService := service.NewLedgerService(store, userClient, merchantClient)
	ledgerHandler := handler.NewLedgerHandler(ledgerService)

	router := gin.Default()
	routes.Setup(router, ledgerHandler, cfg.JWTSecret)

	log.Printf(
		"ledger-service listening on :%s (db=%s users=%s merchants=%s)",
		cfg.ServerPort,
		cfg.DBName,
		cfg.UserServiceURL,
		cfg.MerchantServiceURL,
	)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}
