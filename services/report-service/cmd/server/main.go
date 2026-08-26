package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"paylater/services/report-service/internal/client"
	"paylater/services/report-service/internal/config"
	"paylater/services/report-service/internal/handler"
	"paylater/services/report-service/internal/routes"
	"paylater/services/report-service/internal/service"
)

func main() {
	cfg := config.Load()
	if cfg.InternalAPIToken == "" {
		log.Fatal("INTERNAL_API_TOKEN is required")
	}

	users := client.NewUserClient(cfg.UserServiceURL, cfg.InternalAPIToken)
	ledger := client.NewLedgerClient(cfg.LedgerServiceURL, cfg.InternalAPIToken)
	reportService := service.NewReportService(users, ledger)
	reportHandler := handler.NewReportHandler(reportService)

	router := gin.Default()
	routes.Setup(router, reportHandler, cfg.JWTSecret)

	log.Printf(
		"report-service listening on :%s (users=%s ledger=%s)",
		cfg.ServerPort,
		cfg.UserServiceURL,
		cfg.LedgerServiceURL,
	)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}
