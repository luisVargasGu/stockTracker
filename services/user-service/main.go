package userservice

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/luisVargasGu/stockTracker/common/middleware"
)

func main() {
	tokenService := middleware.NewTokenService(os.Getenv("JWT_SECRET"))

	r := gin.New()

	// Users Auth now
	r.Use(middleware.AuthMiddleware(tokenService))
	fmt.Printf(
		"started",
	)
}
