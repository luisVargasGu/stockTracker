package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/luisVargasGu/stockTracker/user-service/models"
	"go.uber.org/zap"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type Store struct {
	db  *sqlx.DB
	log *zap.Logger
}

func NewStore(db *sqlx.DB, logger *zap.Logger) *Store {
	return &Store{db: db, log: logger}
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
func (s *Store) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	// Check if user already exists
	var count int
	err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM Users WHERE Username = $1", user.Username)
	if err != nil {
		s.log.Error("Error querying database for existing user", zap.Error(err))
		return nil, err
	}

	if count > 0 {
		s.log.Warn("User already exists", zap.String("username", user.Username))
		return nil, ErrUserAlreadyExists
	}

	// Insert new user
	err = s.db.QueryRowxContext(ctx, createUserQuery, user).Scan(&user.ID)
	if err != nil {
		s.log.Error("Error creating user", zap.String("username", user.Username), zap.Error(err))
		return nil, err
	}

	s.log.Info("User created successfully", zap.Int("userID", user.ID), zap.String("username", user.Username))
	return user, nil
}

// GetUserByEmail retrieves a user by their email.
func (s *Store) GetUserByEmail(ctx context.Context, username string) (*models.User, error) {
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
func (s *Store) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
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
func (s *Store) GetUsers(ctx context.Context, offset, limit int) ([]*models.User, error) {
	const query = getUserByBase + `
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var users []*models.User
	err := s.db.SelectContext(ctx, &users, query)
	if err != nil {
		s.log.Error("Error querying users", zap.Int("offset", offset), zap.Int("limit", limit))
		return nil, err
	}

	s.log.Info("Users retrieved successfully", zap.Int("count", len(users)), zap.Int("offset", offset), zap.Int("limit", limit))
	return users, nil
}

func (s *Store) UpdateUser(ctx context.Context, id int, updates map[string]interface{}) (*models.User, error) {
	logger := s.log.With(zap.Int("ID", id))

	// Validate input
	if len(updates) == 0 {
		logger.Warn("No updates provided")
		return nil, errors.New("no updates provided")
	}

	// Strict field validation
	validFields := map[string]bool{
		"username": true,
		"avatar":   true,
		"name":     true,
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
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		logger.Error("Failed to update user", zap.Error(err))
		return nil, err
	}

	return &updatedUser, nil
}

func (s *Store) DeleteUser(ctx context.Context, id int) (bool, error) {
	logger := s.log.With(zap.Int("ID", id))

	query := "DELETE FROM Users WHERE id = $1"
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		return false, fmt.Errorf("failed to delete user with id %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", zap.Error(err))
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn("No user found to delete")
		return false, nil
	}

	logger.Info("User successfully deleted")
	return true, nil
}
