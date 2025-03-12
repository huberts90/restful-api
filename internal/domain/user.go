package domain

import (
	"time"
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
