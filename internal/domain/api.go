package domain

import (
	"github.com/go-playground/validator/v10"
	"time"
)

// UserCreate represents the data needed to create a new user
type UserCreate struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required,alpha"`
	LastName  string `json:"last_name" validate:"required,alpha"`
}

func (u UserCreate) Validate() error {
	return validator.New().Struct(u)
}

type UserCreateResponse struct {
	ID int64 `json:"id"`
}

// UserUpdate represents the data that can be updated for a user
type UserUpdate struct {
	Email     string `json:"email,omitempty" validate:"omitempty,email"`
	FirstName string `json:"first_name,omitempty" validate:"omitempty,alpha,min=2"`
	LastName  string `json:"last_name,omitempty" validate:"omitempty,alpha,min=2"`
}

func (u UserUpdate) Validate() error {
	return validator.New().Struct(u)
}

// UserResponse represents the data sent back to the client
type UserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts a User to a UserResponse
// @MENTION_ME: do not return data directly from the database
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// PaginatedUsersResponse represents the paginated list of users
// Includes metadata for client-side pagination handling
type PaginatedUsersResponse struct {
	Users      []UserResponse `json:"users"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}
