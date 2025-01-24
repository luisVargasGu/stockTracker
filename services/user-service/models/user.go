package models

import (
	"context"
	"time"
)

type User struct {
	ID           int        `json:"id" db:"id"`
	Name         string     `json:"name" db:"name"`
	Username     string     `json:"email" db:"username"`
	Role         string     `json:"role" db:"role"` // e.g., "admin", "user"
	PasswordHash string     `json:"-" db:"password_hash"`
	Avatar       []byte     `json:"avatar,omitempty" db:"avatar"`
	LastLogin    time.Time  `json:"lastLogin" db:"last_login"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	DeletedAt    *time.Time `json:"deletedAt,omitempty" db:"deleted_at"`
}

type UserInfo struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatarUrl,omitempty"`
}

type UserStore interface {
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id int) (*User, error)
	GetUsers(ctx context.Context, offset, limit int) ([]*User, error)
	CreateUser(ctx context.Context, user *User) (*User, error)
	UpdateUser(ctx context.Context, id int, updates map[string]interface{}) (*User, error)
	DeleteUser(ctx context.Context, id int) (bool, error)
}

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Avatar   string `json:"avatar,omitempty"` // URL or base64 encoded
}

type LoginResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	User    *UserInfo  `json:"user,omitempty"`
	Token   *AuthToken `json:"token,omitempty"` // For authentication
}

type AuthToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}
