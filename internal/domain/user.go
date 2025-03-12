package domain

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// @MENTION_ME: field alignment

// User represents the user entity in our system
// Using proper field tags for JSON serialization and database mapping
type User struct {
	ID        int64
	Email     string
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserCreate represents the data needed to create a new user
type UserCreate struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}

// UserUpdate represents the data that can be updated for a user
type UserUpdate struct {
	Email     string `json:"email,omitempty" validate:"omitempty,email"`
	FirstName string `json:"first_name,omitempty" validate:"omitempty,min=2"`
	LastName  string `json:"last_name,omitempty" validate:"omitempty,min=2"`
}

type UserCreateResponse struct {
	ID int64 `json:"id"`
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

// Validate validates the UserCreate struct
// Ensures all required fields are present and valid
func (u UserCreate) Validate() error {
	return validator.New().Struct(u)
}

// Validate validates the UserUpdate struct
// Ensures all provided fields are valid
func (u UserUpdate) Validate() error {
	return validator.New().Struct(u)
}
