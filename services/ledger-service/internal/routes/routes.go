package routes

import (
	"github.com/gin-gonic/gin"

	"paylater/services/ledger-service/internal/handler"
	ledgermw "paylater/services/ledger-service/internal/middleware"
	"paylater/shared/middleware"
	"paylater/shared/response"
)

// Setup registers ledger HTTP routes.
func Setup(router *gin.Engine, h *handler.LedgerHandler, jwtSecret string) {
	router.GET("/health", func(c *gin.Context) {
		response.JSON(c, 200, gin.H{
			"status":  "ok",
			"service": "ledger-service",
		})
	})

	user := router.Group("/")
	user.Use(
		middleware.AuthMiddleware(jwtSecret),
		middleware.RequireRole("user"),
	)
	user.POST("/purchases", h.Purchase)
	user.POST("/payments", h.Repay)
	user.GET("/payments", h.ListMyPayments)

	admin := router.Group("/admin")
	admin.Use(
		middleware.AuthMiddleware(jwtSecret),
		middleware.RequireRole("admin"),
	)
	admin.GET("/purchases", h.ListPurchases)
	admin.GET("/purchases/:id", h.GetPurchaseByID)
	admin.GET("/users/:id/purchases", h.ListUserPurchases)
	admin.GET("/payments/:id", h.GetPaymentByID)
	admin.GET("/users/:id/payments", h.ListUserPaymentsAdmin)

	merchant := router.Group("/merchant")
	merchant.Use(
		middleware.AuthMiddleware(jwtSecret),
		middleware.RequireRole("merchant"),
	)
	merchant.GET("/transactions", h.ListMerchantTransactions)

	internal := router.Group("/internal")
	internal.Use(ledgermw.RequireInternalToken())
	internal.GET("/reports/merchant-commissions", h.MerchantCommissions)
}
