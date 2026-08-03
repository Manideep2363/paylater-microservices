package routes

import (
	"github.com/gin-gonic/gin"

	"paylater/services/merchant-service/internal/handler"
	merchantmw "paylater/services/merchant-service/internal/middleware"
	"paylater/shared/middleware"
	"paylater/shared/response"
)

// Setup registers health, merchant, admin, and internal routes.
func Setup(router *gin.Engine, h *handler.MerchantHandler, jwtSecret string) {
	router.GET("/health", func(c *gin.Context) {
		response.JSON(c, 200, gin.H{
			"status":  "ok",
			"service": "merchant-service",
		})
	})

	merchant := router.Group("/merchant")
	merchant.Use(
		middleware.AuthMiddleware(jwtSecret),
		middleware.RequireRole("merchant"),
	)
	merchant.GET("/profile", h.GetProfile)

	admin := router.Group("/admin")
	admin.Use(
		middleware.AuthMiddleware(jwtSecret),
		middleware.RequireRole("admin"),
	)
	admin.POST("/merchants", h.CreateMerchant)
	admin.GET("/merchants", h.ListMerchants)
	admin.GET("/merchants/:id", h.GetMerchantByID)
	admin.PUT("/merchants/:id/commission", h.UpdateMerchantCommission)

	internal := router.Group("/internal")
	internal.Use(merchantmw.RequireInternalToken())
	internal.POST("/merchants", h.InternalCreateMerchant)
	internal.GET("/merchants/by-email", h.InternalGetMerchantByEmail)
	internal.GET("/merchants/:id", h.InternalGetMerchantByID)
}
