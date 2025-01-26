package services

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luisVargasGu/stockTracker/common/middleware"
	"github.com/luisVargasGu/stockTracker/user-service/models"
	"go.uber.org/zap"
)

var (
	AdminRole              = "admin"
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInternalServerError = errors.New("internal server error")
	ErrUnauthorized        = errors.New("user is unauthorized")
	ErrDuplicateEmail      = errors.New("email is a duplicate")
	ErrInvalidEmail        = errors.New("invalid email")
	ErrWeakPassword        = errors.New("weak password")
	ErrTokenGeneration     = errors.New("token generation failed")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserDeleted         = errors.New("user already deleted")
)

type UserService struct {
	repo         models.UserRepository
	tokenService middleware.TokenService
	log          *zap.Logger
}

func NewUserService(repository models.UserRepository,
	tokenService middleware.TokenService,
	log *zap.Logger) *UserService {
	return &UserService{repo: repository, tokenService: tokenService, log: log}
}

func (s *UserService) LoginUser(ctx *gin.Context, payload models.LoginUserPayload) (*models.LoginResponse, error) {
	// Prepare default response
	response := &models.LoginResponse{
		Success: false,
		Message: "Login failed",
	}

	// Find user by email
	user, err := s.repo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.Message = ErrInvalidCredentials.Error()
			return response, ErrInvalidCredentials
		}
		s.log.Error("Unexpected error fetching user", zap.Error(err))
		response.Message = ErrInternalServerError.Error()
		return response, err
	}

	// Verify password
	if !middleware.ComparePasswords(user.PasswordHash, []byte(payload.Password)) {
		response.Message = ErrInvalidCredentials.Error()
		return response, ErrInvalidCredentials
	}

	// Check user status
	if user.DeletedAt != nil {
		response.Message = ErrUserDeleted.Error()
		return response, ErrUserDeleted
	}

	// Generate token pair
	accessToken, err := s.tokenService.GenerateToken(
		strconv.Itoa(user.ID),
		user.Email,
	)
	if err != nil {
		response.Message = ErrTokenGeneration.Error()
		s.log.Error("Error generating tokens", zap.Error(err))
		return response, err
	}

	// Update last login (non-blocking)
	go func() {
		if _, err := s.repo.UpdateUser(context.Background(),
			user.ID, map[string]interface{}{"updated_at": time.Now()}); err != nil {
			s.log.Error("Failed to update last login", zap.Error(err))
		}
	}()

	// Prepare response
	response.Success = true
	response.Message = "Login successful"
	response.User = &models.UserInfo{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
	}
	response.Token = &models.AuthToken{
		Token:     accessToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	return response, nil
}

func (s *UserService) RegisterUser(ctx *gin.Context, payload models.RegisterUserPayload) (*models.User, error) {
	// Check if user already exists
	existingUser, err := s.repo.GetUserByEmail(ctx, payload.Email)
	if err != nil && err != ErrUserNotFound {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := middleware.HashPassword(payload.Password)
	if err != nil {
		return nil, err
	}

	// Create user model
	newUser := &models.User{
		Email:        payload.Email,
		Name:         payload.Name,
		PasswordHash: hashedPassword,
		Avatar:       []byte(payload.Avatar),
	}

	// Save user to repository
	return s.repo.CreateUser(ctx, newUser)
}

func (s *UserService) GetUsers(ctx context.Context, offset, limit int) ([]*models.User, int, error) {
	// Fetch paginated users
	users, totalUsers, err := s.repo.GetUsers(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return users, totalUsers, nil
}

func (s *UserService) GetUserByID(ctx *gin.Context, id int) (*models.User, error) {
	currentUser, err := s.checkPermissions(ctx, id)
	if err != nil {
		return nil, err
	}

	return currentUser, nil
}

func (s *UserService) UpdateUser(ctx *gin.Context, id int, updates map[string]interface{}) (*models.User, error) {
	_, err := s.checkPermissions(ctx, id)
	if err != nil {
		return nil, err
	}

	updatedUser, err := s.repo.UpdateUser(ctx, id, updates)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s *UserService) hasAccessToUser(user *models.User, id int) bool {
	return user.ID == id || user.Role == AdminRole
}

func (s *UserService) DeleteUser(ctx *gin.Context, id int) error {
	currentUser, err := s.extractUserFromContext(ctx)
	if err != nil {
		return ErrUnauthorized
	}

	// Check if user has permission
	if !s.hasAccessToUser(currentUser, id) {
		return ErrUnauthorized
	}

	err = s.repo.DeleteUser(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) checkPermissions(ctx *gin.Context, id int) (*models.User, error) {
	// Authenticate/authorize user from context
	currentUser, err := s.extractUserFromContext(ctx)
	if err != nil {
		return nil, ErrUnauthorized
	}

	// Check if user has permission
	if !s.hasAccessToUser(currentUser, id) {
		return nil, ErrUnauthorized
	}

	return currentUser, nil
}

func (s *UserService) extractUserFromContext(ctx *gin.Context) (*models.User, error) {
	// Extract user ID from Gin context
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, ErrUnauthorized
	}

	// Convert string to int if needed
	id, err := strconv.Atoi(userID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	// Fetch user
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, ErrUnauthorized
	}

	return user, nil
}

func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}
