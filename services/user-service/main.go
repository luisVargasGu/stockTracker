package main

import (
	"log"
	"os"

	"github.com/luisVargasGu/stockTracker/common/middleware"
	"github.com/luisVargasGu/stockTracker/user-service/api"
	"github.com/luisVargasGu/stockTracker/user-service/db"
	"go.uber.org/zap"
)

func main() {
	tokenService := middleware.NewTokenService(os.Getenv("JWT_SECRET"))

	logger, err := zap.NewDevelopment() // For production: zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	db := db.DbConnect(logger)
	server := api.NewAPIServer(":8080", db)
	server.Run(logger, *tokenService)
}
