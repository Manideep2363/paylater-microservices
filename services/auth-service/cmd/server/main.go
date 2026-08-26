package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"paylater/services/auth-service/internal/client"
	"paylater/services/auth-service/internal/config"
	"paylater/services/auth-service/internal/handler"
	"paylater/services/auth-service/internal/routes"
	"paylater/services/auth-service/internal/service"
)

func main() {
	cfg := config.Load()

	if cfg.InternalAPIToken == "" {
		log.Fatal("INTERNAL_API_TOKEN is required for user/merchant service calls")
	}

	userRepo := client.NewUserClient(cfg.UserServiceURL, cfg.InternalAPIToken)
	merchantRepo := client.NewMerchantClient(cfg.MerchantServiceURL, cfg.InternalAPIToken)

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

	log.Printf(
		"auth-service listening on :%s (users=%s merchants=%s)",
		cfg.ServerPort,
		cfg.UserServiceURL,
		cfg.MerchantServiceURL,
	)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}
