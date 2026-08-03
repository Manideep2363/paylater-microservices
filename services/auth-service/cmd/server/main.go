package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"paylater/shared/config"
	"paylater/shared/response"
)

func main() {
	cfg := config.LoadConfig()

	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		response.JSON(c, 200, gin.H{
			"status":  "ok",
			"service": "auth-service",
		})
	})

	log.Printf("auth-service listening on :%s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}
