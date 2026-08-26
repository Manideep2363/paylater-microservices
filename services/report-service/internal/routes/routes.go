package routes

import (
	"github.com/gin-gonic/gin"

	"paylater/services/report-service/internal/handler"
	"paylater/shared/middleware"
	"paylater/shared/response"
)

// Setup registers report-service routes.
func Setup(router *gin.Engine, h *handler.ReportHandler, jwtSecret string) {
	router.GET("/health", func(c *gin.Context) {
		response.JSON(c, 200, gin.H{
			"status":  "ok",
			"service": "report-service",
		})
	})

	admin := router.Group("/admin/reports")
	admin.Use(
		middleware.AuthMiddleware(jwtSecret),
		middleware.RequireRole("admin"),
	)
	admin.GET("/outstanding-balance", h.OutstandingBalance)
	admin.GET("/users-due", h.UsersDue)
	admin.GET("/users-at-credit-limit", h.UsersAtCreditLimit)
	admin.GET("/merchant-commissions", h.MerchantCommissions)
}
