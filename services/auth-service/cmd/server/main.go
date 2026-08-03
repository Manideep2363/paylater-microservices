package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"paylater/services/auth-service/internal/handler"
	"paylater/services/auth-service/internal/repository/memory"
	"paylater/services/auth-service/internal/routes"
	"paylater/services/auth-service/internal/service"
	"paylater/shared/config"
)

func main() {
	cfg := config.LoadConfig()

	// Temporary in-memory adapters standing in for future user/merchant REST clients.
	userRepo := memory.NewUserStore()
	merchantRepo := memory.NewMerchantStore()

	authService := service.NewAuthService(
		userRepo,
		merchantRepo,
		cfg.JWTSecret,
		cfg.AdminEmail,
		cfg.AdminPassword,
	)
	authHandler := handler.NewAuthHandler(authService)

	router := gin.Default()
	routes.Setup(router, authHandler)

	log.Printf("auth-service listening on :%s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}
