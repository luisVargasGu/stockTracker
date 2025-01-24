package models

import (
	"time"
)

type User struct {
	ID           int        `json:"id"`
	Name         string     `json:"name"`
	Username     string     `json:"email"`
	Role         string     `json:"role"` // e.g., "admin", "user"
	PasswordHash string     `json:"-"`
	Avatar       []byte     `json:"avatar,omitempty"`
	LastLogin    time.Time  `json:"lastLogin"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	DeletedAt    *time.Time `json:"deletedAt,omitempty"`
}

type UserInfo struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatarUrl,omitempty"`
}

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
	GetUsers(offset, limit int) ([]*User, error)
	CreateUser(user *User) (*User, error)
	UpdateUser(id int, updates map[string]interface{}) (*User, error)
	DeleteUser(id int) (int, error)
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
