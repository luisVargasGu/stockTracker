package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/luisVargasGu/stockTracker/common/middleware"
	"github.com/luisVargasGu/stockTracker/user-service/controllers"
	"github.com/luisVargasGu/stockTracker/user-service/repository"
	"github.com/luisVargasGu/stockTracker/user-service/services"
	"go.uber.org/zap"
)

type APIServer struct {
	addr string
	db   *sqlx.DB
}

func NewAPIServer(addr string, db *sqlx.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run(logger *zap.Logger, tokenService middleware.TokenService) error {
	router := gin.New()

	router.Use(middleware.LoggingMiddleware(logger))

	router.Use(gin.Recovery()) // Ensures server recovers from panics

	// Initialize dependencies
	userRepository := repository.NewUserStore(s.db, logger)
	userService := services.NewUserService(userRepository, tokenService, logger)
	userHandler := controllers.NewUserHandler(userService, tokenService, logger)
	userHandler.RegisterRoutes(router.Group("/api/v1"))

	// Add a health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	// Start the server with error handling
	logger.Info("Starting server on ", zap.String("adders", s.addr))
	if err := router.Run(s.addr); err != nil {
		logger.Info("Server failed to start", zap.Error(err))
		return err
	}

	return nil
}
