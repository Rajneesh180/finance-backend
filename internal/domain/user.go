package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleAnalyst Role = "analyst"
	RoleViewer  Role = "viewer"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleAnalyst, RoleViewer:
		return true
	}
	return false
}

type User struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Password  string     `json:"-"`
	Name      string     `json:"name"`
	Role      Role       `json:"role"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Name     string `json:"name" validate:"required,max=255"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UpdateProfileRequest struct {
	Email string `json:"email" validate:"omitempty,email,max=255"`
	Name  string `json:"name" validate:"omitempty,max=255"`
}

type AdminUpdateUserRequest struct {
	Role     *Role `json:"role" validate:"omitempty"`
	IsActive *bool `json:"is_active" validate:"omitempty"`
}
