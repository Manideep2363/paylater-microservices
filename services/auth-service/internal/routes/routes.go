package routes

import (
	"github.com/gin-gonic/gin"

	"paylater/services/auth-service/internal/handler"
	"paylater/shared/response"
)

// SetupRegisters public auth routes and health check.
func Setup(router *gin.Engine, authHandler *handler.AuthHandler) {
	router.GET("/health", func(c *gin.Context) {
		response.JSON(c, 200, gin.H{
			"status":  "ok",
			"service": "auth-service",
		})
	})

	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)
	router.POST("/merchant/register", authHandler.MerchantRegister)
	router.POST("/merchant/login", authHandler.MerchantLogin)
	router.POST("/admin/login", authHandler.AdminLogin)
}
