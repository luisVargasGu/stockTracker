package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/luisVargasGu/stockTracker/common/middleware"
	"github.com/luisVargasGu/stockTracker/common/utils"
	"github.com/luisVargasGu/stockTracker/user-service/models"
	"github.com/luisVargasGu/stockTracker/user-service/services"
	"go.uber.org/zap"
)

type UserHandler struct {
	service models.UserService
	ts      *middleware.TokenService
	log     *zap.Logger
}

func NewUserHandler(service models.UserService, tokenService *middleware.TokenService, log *zap.Logger) *UserHandler {
	return &UserHandler{service: service, ts: tokenService, log: log}
}

func (h *UserHandler) RegisterRoutes(r *gin.Engine) {
	r.OPTIONS("/users/register", middleware.CorsMiddleware())
	r.POST("/users/register", middleware.CorsMiddleware(), h.RegisterUser)

	r.OPTIONS("/users/login", middleware.CorsMiddleware())
	r.POST("/users/login", middleware.CorsMiddleware(), h.LoginUser)

	r.OPTIONS("/users/:id", middleware.CorsMiddleware())
	r.GET("/users/:id",
		middleware.CorsMiddleware(),
		middleware.AuthMiddleware(h.ts),
		h.GetUserByID)

	r.OPTIONS("/users/:id", middleware.CorsMiddleware())
	r.PUT("/users/:id",
		middleware.CorsMiddleware(),
		middleware.AuthMiddleware(h.ts),
		h.UpdateUser)

	r.OPTIONS("/users/:id", middleware.CorsMiddleware())
	r.DELETE("/users/:id",
		middleware.CorsMiddleware(),
		middleware.AuthMiddleware(h.ts),
		h.DeleteUser)

	r.OPTIONS("/users", middleware.CorsMiddleware())
	r.GET("/users",
		middleware.CorsMiddleware(),
		middleware.AuthMiddleware(h.ts),
		h.GetUsers)
}

// RegisterUser handles user registration
func (h *UserHandler) RegisterUser(c *gin.Context) {
	var payload models.RegisterUserPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		h.log.Error("Validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("validation failed: %w", err)))
		return
	}

	// Call the service layer to handle registration logic
	user, err := h.service.RegisterUser(c, payload)
	if err != nil {
		h.log.Error("Failed to register user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errorResponse(errors.New("Failed to register user")))
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// LoginUser handles user login
func (h *UserHandler) LoginUser(c *gin.Context) {
	var payload models.LoginUserPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	validate := validator.New()
	if err := validate.Struct(payload); err != nil {
		h.log.Error("Validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("validation failed: %w", err)))
		return
	}

	// Call the service layer to handle login logic
	loginResponse, err := h.service.LoginUser(c, payload)
	if err != nil {
		h.log.Error("Failed to login user", zap.Error(err))
		c.JSON(http.StatusUnauthorized, errorResponse(services.ErrUnauthorized))
		return
	}

	c.JSON(http.StatusOK, loginResponse)
}

// GetUserByID retrieves user details by ID
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id, err := h.parseUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Call the service layer to get user details
	user, err := h.service.GetUserByID(c, id)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, errorResponse(err))
		case errors.Is(err, services.ErrUnauthorized):
			c.JSON(http.StatusForbidden, errorResponse(err))
		default:
			h.log.Error("Failed to get user",
				zap.Int("userID", id),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, errorResponse(errors.New("failed to retrieve user")))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// UpdateUser handles updating user details
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := h.parseUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req models.UpdateUserPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	// Validate request using validator
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		h.log.Error("Validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("validation failed: %w", err)))
		return
	}

	// Convert strongly typed request to map for service layer
	updates := make(map[string]interface{})
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}

	// Call the service layer to update user details
	updatedUser, err := h.service.UpdateUser(c, id, updates)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, errorResponse(err))
		case errors.Is(err, services.ErrDuplicateEmail):
			c.JSON(http.StatusConflict, errorResponse(err))
		default:
			h.log.Error("Failed to update user",
				zap.Int("userID", id),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, errorResponse(errors.New("update failed")))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": updatedUser})
}

// DeleteUser handles deleting a user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := h.parseUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := h.service.DeleteUser(c, id); err != nil {
		// Use a more robust error handling approach
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, errorResponse(err))
		case errors.Is(err, services.ErrUnauthorized):
			c.JSON(http.StatusForbidden, errorResponse(err))
		default:
			h.log.Error("Failed to delete user",
				zap.Int("userID", id),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, errorResponse(errors.New("failed to delete user")))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// GetUsers retrieves a list of users with pagination
func (h *UserHandler) GetUsers(c *gin.Context) {
	// Centralize pagination parameter parsing
	pagination, err := h.parsePaginationParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
	}

	// Get users with the validated offset and limit
	users, total, err := h.service.GetUsers(c, pagination.Offset, pagination.Limit)
	if err != nil {
		h.log.Error("Failed to fetch users",
			zap.Int("offset", pagination.Offset),
			zap.Int("limit", pagination.Limit),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Return the list of users and metadata
	c.JSON(http.StatusOK, gin.H{
		"users":  users,
		"total":  total,
		"offset": pagination.Offset,
		"limit":  pagination.Limit,
	})
}

// Parse and validate pagination parameters
func (h *UserHandler) parsePaginationParams(c *gin.Context) (models.Pagination, error) {
	offset, err := utils.ConvertQueryParamToInt(c, "offset", 0, 0, 10000)
	if err != nil {
		return models.Pagination{}, fmt.Errorf("invalid offset: %w", err)
	}

	limit, err := utils.ConvertQueryParamToInt(c, "limit", 10, 1, 1000)
	if err != nil {
		return models.Pagination{}, fmt.Errorf("invalid limit: %w", err)
	}

	return models.Pagination{
		Offset: offset,
		Limit:  limit,
	}, nil
}

// Helper method to parse and validate user ID
func (h *UserHandler) parseUserID(c *gin.Context) (int, error) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID format: %w", err)
	}
	return id, nil
}

// Centralized error response helper
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
