package storage

import (
	"context"

	"github.com/huberts90/restful-api/internal/domain"
)

// Storer defines the contract for all storage implementations
// This interface decouples business logic from the specific database used
// @MENTION_ME
type Storer interface {
	CreateUser(ctx context.Context, user domain.UserCreate) (int64, error)
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	UpdateUser(ctx context.Context, id int64, user domain.UserUpdate) error
	DeleteUser(ctx context.Context, id int64) error
	ListUsers(ctx context.Context, page, pageSize int) ([]domain.User, int, error)
	Close() error
}
