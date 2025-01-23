package userservice

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/luisVargasGu/stockTracker/common/middleware"
	"go.uber.org/zap"
)

func main() {
	tokenService := middleware.NewTokenService(os.Getenv("JWT_SECRET"))

	logger, err := zap.NewDevelopment() // For production: zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	r := gin.New()

	r.Use(middleware.LoggingMiddleware(logger))

	// Routes after this requre auth
	r.Use(middleware.AuthMiddleware(tokenService))

	// TODO: When registering routes with handlers pass the logger
	fmt.Printf(
		"started",
	)
}
