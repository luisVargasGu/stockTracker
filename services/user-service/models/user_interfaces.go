package models

import "context"

type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id int) (*User, error)
	GetUsers(ctx context.Context, offset, limit int) ([]*User, int, error)
	CreateUser(ctx context.Context, user *User) (*User, error)
	UpdateUser(ctx context.Context, id int, updates map[string]interface{}) (*User, error)
	DeleteUser(ctx context.Context, id int) (bool, error)
}

type UserService interface {
	GetUsers(ctx context.Context, offset, limit int) ([]*User, int, error)
	LoginUser(ctx context.Context, user LoginUserPayload) (*LoginResponse, error)
	GetUserByID(ctx context.Context, id int) (*User, error)
	RegisterUser(ctx context.Context, user RegisterUserPayload) (*User, error)
	UpdateUser(ctx context.Context, id int, updates map[string]interface{}) (*User, error)
	DeleteUser(ctx context.Context, id int) error
}
