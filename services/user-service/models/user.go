package models

import (
	"time"
)

type User struct {
	ID           int        `json:"id" db:"id"`
	Name         string     `json:"name" db:"name"`
	Email        string     `json:"email" validate:"email" db:"username"`
	Role         string     `json:"role" db:"role"` // e.g., "admin", "user"
	PasswordHash string     `json:"-" db:"password_hash"`
	Avatar       []byte     `json:"avatar,omitempty" db:"avatar"`
	LastLogin    time.Time  `json:"lastLogin" db:"last_login"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	DeletedAt    *time.Time `json:"deletedAt,omitempty" db:"deleted_at"`
}

// TODO: may not need this
type UserInfo struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatarUrl,omitempty"`
}
