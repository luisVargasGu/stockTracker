package models

import "time"

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name,omitempty"`
	Password string `json:"password" validate:"required,min=8"`
	Avatar   string `json:"avatar,omitempty"` // URL or base64 encoded
}

type LoginResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	User    *UserInfo  `json:"user,omitempty"`
	Token   *AuthToken `json:"token,omitempty"` // For authentication
}

type UpdateUserPayload struct {
	Name   *string `json:"name,omitempty"`
	Email  *string `json:"email,omitemty" validate:"email"`
	Avatar *[]byte `json:"avatar,omitempty"`
}

type Pagination struct {
	Offset int
	Limit  int
}

type AuthToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}
