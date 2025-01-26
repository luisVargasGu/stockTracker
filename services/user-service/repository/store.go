package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/luisVargasGu/stockTracker/user-service/models"
	"github.com/luisVargasGu/stockTracker/user-service/services"
	"go.uber.org/zap"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserStore struct {
	db  *sqlx.DB
	log *zap.Logger
}

func NewUserStore(db *sqlx.DB, logger *zap.Logger) *UserStore {
	return &UserStore{db: db, log: logger}
}

const (
	createUserQuery = `INSERT INTO Users 
		(name, username, role, password_hash, avatar, last_login, updated_at, created_at) 
		VALUES (:name, :username, :role, :password_hash, :avatar, :last_login, :updated_at, :created_at) 
		RETURNING ID`
	allUserFields       = "id, name, username, role, avatar, last_login, updated_at, created_at, deleted_at"
	getUserByBase       = "SELECT " + allUserFields + " FROM Users "
	getUserByEmailQuery = getUserByBase + "WHERE username = $1"
	getUserByIDQuery    = getUserByBase + "WHERE id = $1"
)

// CreateUser inserts a new user into the database.
func (s *UserStore) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	// Check if user already exists
	var count int
	err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM Users WHERE Email = $1", user.Email)
	if err != nil {
		s.log.Error("Error querying database for existing user", zap.Error(err))
		return nil, err
	}

	if count > 0 {
		s.log.Warn("User already exists", zap.String("username", user.Email))
		return nil, ErrUserAlreadyExists
	}

	// Insert new user
	err = s.db.QueryRowxContext(ctx, createUserQuery, user).Scan(&user.ID)
	if err != nil {
		s.log.Error("Error creating user", zap.String("username", user.Email), zap.Error(err))
		return nil, err
	}

	s.log.Info("User created successfully", zap.Int("userID", user.ID), zap.String("username", user.Email))
	return user, nil
}

// GetUserByEmail retrieves a user by their email.
func (s *UserStore) GetUserByEmail(ctx context.Context, username string) (*models.User, error) {
	var user models.User

	err := s.db.GetContext(ctx, &user, getUserByEmailQuery, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Warn("User not found", zap.String("username", username))
			return nil, ErrUserNotFound
		}
		s.log.Error("Error querying user by ID", zap.String("username", username), zap.Error(err))
		return nil, err
	}

	s.log.Info("User retrieved successfully", zap.String("username", username))
	return &user, nil
}

// GetUserByID retrieves a user by their ID.
func (s *UserStore) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	var user models.User

	err := s.db.GetContext(ctx, &user, getUserByIDQuery, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Warn("User not found", zap.Int("userID", userID))
			return nil, ErrUserNotFound
		}
		s.log.Error("Error querying user by ID", zap.Int("userID", userID), zap.Error(err))
		return nil, err
	}

	s.log.Info("User retrieved successfully", zap.Int("userID", userID))
	return &user, nil
}

// GetUsers retrieves multiple users
func (s *UserStore) GetUsers(ctx context.Context, offset, limit int) ([]*models.User, int, error) {
	const query = getUserByBase + `
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var users []*models.User
	err := s.db.SelectContext(ctx, &users, query)
	if err != nil {
		s.log.Error("Error querying users", zap.Int("offset", offset), zap.Int("limit", limit))
		return nil, 0, err
	}

	s.log.Info("Users retrieved successfully", zap.Int("count", len(users)), zap.Int("offset", offset), zap.Int("limit", limit))
	return users, len(users), nil
}

func (s *UserStore) UpdateUser(ctx context.Context, id int, updates map[string]interface{}) (*models.User, error) {
	logger := s.log.With(zap.Int("ID", id))

	// Validate input
	if len(updates) == 0 {
		logger.Warn("No updates provided")
		return nil, errors.New("no updates provided")
	}

	// Strict field validation
	validFields := map[string]bool{
		"email":      true,
		"username":   true,
		"updated_at": true,
		"avatar":     true,
		"name":       true,
	}

	// Prepare query builder
	var (
		queryBuilder strings.Builder
		args         []interface{}
		i            = 1
	)
	queryBuilder.WriteString("UPDATE Users SET ")

	for key, value := range updates {
		if !validFields[key] {
			logger.Warn("Invalid update field", zap.String("key", key))
			return nil, fmt.Errorf("invalid update field: %s", key)
		}

		if i > 1 {
			queryBuilder.WriteString(", ")
		}
		queryBuilder.WriteString(fmt.Sprintf("%s = $%d", key, i))
		args = append(args, value)
		i++
	}

	queryBuilder.WriteString(fmt.Sprintf(" WHERE id = $%d RETURNING "+allUserFields, i))
	args = append(args, id)

	// Use QueryRowContext for single row return
	var updatedUser models.User
	err := s.db.GetContext(ctx, &updatedUser, queryBuilder.String(), args...)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("No user found for update", zap.Int("ID", id))
			return nil, services.ErrUserNotFound
		}

		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if strings.Contains(pqErr.Detail, "email") {
					logger.Warn(`Duplicate email,
						user with that email already exists`,
						zap.String("email", updates["email"].(string)))
					return nil, services.ErrDuplicateEmail
				}
			}
		}
		return nil, err
	}

	return &updatedUser, nil
}

func (s *UserStore) DeleteUser(ctx context.Context, id int) error {
	logger := s.log.With(zap.Int("ID", id))

	query := "DELETE FROM Users WHERE id = $1"
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		return fmt.Errorf("failed to delete user with id %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn("No user found to delete")
		return services.ErrUserNotFound
	}

	logger.Info("User successfully deleted")
	return nil
}
