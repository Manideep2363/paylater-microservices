package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"paylater/services/user-service/internal/database"
	"paylater/services/user-service/internal/handler"
	"paylater/services/user-service/internal/repository"
	"paylater/services/user-service/internal/routes"
	"paylater/services/user-service/internal/service"
	"paylater/shared/config"
)

func main() {
	cfg := config.LoadConfig()
	applyUserServiceDefaults(cfg)

	dbConn, err := database.NewMySQL(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer dbConn.Close()

	repo := repository.NewSQLRepository(dbConn)
	userService := service.NewUserService(repo)
	userHandler := handler.NewUserHandler(userService)

	router := gin.Default()
	routes.Setup(router, userHandler, cfg.JWTSecret)

	log.Printf("user-service listening on :%s (db=%s)", cfg.ServerPort, cfg.DBName)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}

// applyUserServiceDefaults sets service-specific defaults when env vars are unset.
func applyUserServiceDefaults(cfg *config.Config) {
	if os.Getenv("DB_NAME") == "" {
		cfg.DBName = "paylater_users"
	}
	if os.Getenv("SERVER_PORT") == "" {
		cfg.ServerPort = "8082"
	}
}
