package routes

import (
	"github.com/gin-gonic/gin"

	"paylater/services/user-service/internal/handler"
	usermw "paylater/services/user-service/internal/middleware"
	"paylater/shared/middleware"
	"paylater/shared/response"
)

// Setup registers health, public/admin, and internal user routes.
func Setup(router *gin.Engine, h *handler.UserHandler, jwtSecret string) {
	router.GET("/health", func(c *gin.Context) {
		response.JSON(c, 200, gin.H{
			"status":  "ok",
			"service": "user-service",
		})
	})

	// Authenticated user/admin routes.
	users := router.Group("/")
	users.Use(middleware.AuthMiddleware(jwtSecret))
	users.GET("/users/:id", h.GetUserByID)

	admin := router.Group("/admin")
	admin.Use(
		middleware.AuthMiddleware(jwtSecret),
		middleware.RequireRole("admin"),
	)
	admin.GET("/users", h.ListUsers)
	admin.POST("/users", h.CreateUser)

	// Internal routes for future service-to-service calls.
	internal := router.Group("/internal")
	internal.Use(usermw.RequireInternalToken())
	internal.POST("/users", h.InternalCreateUser)
	internal.GET("/users/by-email", h.InternalGetUserByEmail)
	internal.GET("/users/:id", h.InternalGetUserByID)
	internal.POST("/users/:id/due/increase", h.IncreaseDue)
	internal.POST("/users/:id/due/decrease", h.DecreaseDue)
	internal.GET("/reports/outstanding-balance", h.OutstandingBalance)
	internal.GET("/reports/users-due", h.UserOutstandingDues)
	internal.GET("/reports/users-at-credit-limit", h.UsersAtCreditLimit)
}
