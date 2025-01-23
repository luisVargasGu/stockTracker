package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TODO: may expand with prometheus and grafana
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Generate a unique request ID (correlation ID)
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Writer.Header().Set("X-Request-ID", requestID)

		// Log incoming request
		logger.Info("Incoming request",
			zap.String("method", c.Request.Method),
			zap.String("url", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("request_id", requestID),
		)

		// Process the request
		c.Next()

		// Log response details after processing
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		logger.Info("Completed request",
			zap.String("method", c.Request.Method),
			zap.String("url", c.Request.URL.Path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("request_id", requestID),
		)

		// Capture and log errors if they occurred
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error("Request error",
					zap.String("request_id", requestID),
					zap.String("error", e.Error()),
				)
			}
		}
	}
}
