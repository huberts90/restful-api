package storage

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/huberts90/restful-api/internal/domain"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// testFixture holds test data and utilities
type testFixture struct {
	store    *PostgresStore
	mock     sqlmock.Sqlmock
	now      time.Time
	users    []domain.User
	cleanup  func()
	userRows *sqlmock.Rows
}

// setupTest creates a new test fixture
func setupTest(t *testing.T) *testFixture {
	t.Helper()

	// Use QueryMatcherEqual for exact query matching
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err, "Failed to create mock DB")

	logger := zap.NewNop()
	store := &PostgresStore{
		db:     db,
		logger: logger,
	}

	// Create test data with fixed time to avoid comparison issues
	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	users := []domain.User{
		{
			ID:        1,
			Email:     "user1@example.com",
			FirstName: "John",
			LastName:  "Doe",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        2,
			Email:     "user2@example.com",
			FirstName: "Jane",
			LastName:  "Smith",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// Create reusable user rows
	userRows := sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "created_at", "updated_at"})
	for _, u := range users {
		userRows.AddRow(u.ID, u.Email, u.FirstName, u.LastName, u.CreatedAt, u.UpdatedAt)
	}

	return &testFixture{
		store:    store,
		mock:     mock,
		now:      now,
		users:    users,
		userRows: userRows,
		cleanup: func() {
			mock.ExpectClose()
			db.Close()
		},
	}
}

func TestCreateUser(t *testing.T) {
	f := setupTest(t)
	defer f.cleanup()

	userCreate := domain.UserCreate{
		Email:     "new@example.com",
		FirstName: "New",
		LastName:  "User",
	}

	tests := []struct {
		name    string
		input   domain.UserCreate
		setup   func(sqlmock.Sqlmock)
		want    int64
		wantErr error
	}{
		{
			name:  "success",
			input: userCreate,
			setup: func(mock sqlmock.Sqlmock) {
				// Use the exact SQL query from constants
				mock.ExpectQuery(strings.TrimSpace(sqlCreateUser)).
					WithArgs(userCreate.Email, userCreate.FirstName, userCreate.LastName).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
			},
			want:    int64(3),
			wantErr: nil,
		},
		{
			name:  "duplicate email",
			input: userCreate,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(strings.TrimSpace(sqlCreateUser)).
					WithArgs(userCreate.Email, userCreate.FirstName, userCreate.LastName).
					WillReturnError(&pq.Error{Code: "23505", Constraint: "users_email_key"})
			},
			want:    0,
			wantErr: ErrDuplicateEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(f.mock)

			got, err := f.store.CreateUser(context.Background(), tt.input)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				assert.Zero(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			assert.NoError(t, f.mock.ExpectationsWereMet(), "SQL expectations not met")
		})
	}
}

func TestGetUserByID(t *testing.T) {
	f := setupTest(t)
	defer f.cleanup()

	tests := []struct {
		name    string
		id      int64
		setup   func(sqlmock.Sqlmock)
		want    *domain.User
		wantErr error
	}{
		{
			name: "success",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "created_at", "updated_at"}).
					AddRow(f.users[0].ID, f.users[0].Email, f.users[0].FirstName, f.users[0].LastName, f.users[0].CreatedAt, f.users[0].UpdatedAt)

				mock.ExpectQuery(sqlGetUserByID).
					WithArgs(f.users[0].ID).
					WillReturnRows(rows)
			},
			want:    &f.users[0],
			wantErr: nil,
		},
		{
			name:    "invalid id",
			id:      0,
			setup:   func(mock sqlmock.Sqlmock) {},
			want:    nil,
			wantErr: ErrInvalidID,
		},
		{
			name: "not found",
			id:   999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(sqlGetUserByID).
					WithArgs(int64(999)).
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(f.mock)

			got, err := f.store.GetUserByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			assert.NoError(t, f.mock.ExpectationsWereMet(), "SQL expectations not met")
		})
	}
}

func TestDeleteUser(t *testing.T) {
	f := setupTest(t)
	defer f.cleanup()

	tests := []struct {
		name    string
		id      int64
		setup   func(sqlmock.Sqlmock)
		wantErr error
	}{
		{
			name: "success",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(sqlDeleteUser).
					WithArgs(int64(1)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: nil,
		},
		{
			name:    "invalid id",
			id:      0,
			setup:   func(mock sqlmock.Sqlmock) {},
			wantErr: ErrInvalidID,
		},
		{
			name: "not found",
			id:   999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(sqlDeleteUser).
					WithArgs(int64(999)).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(f.mock)

			err := f.store.DeleteUser(context.Background(), tt.id)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, f.mock.ExpectationsWereMet(), "SQL expectations not met")
		})
	}
}

func TestListUsers(t *testing.T) {
	f := setupTest(t)
	defer f.cleanup()

	tests := []struct {
		name     string
		page     int
		pageSize int
		setup    func(sqlmock.Sqlmock)
		want     []domain.User
		wantN    int
		wantErr  error
	}{
		{
			name:     "success",
			page:     1,
			pageSize: 10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(sqlCountUsers).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(len(f.users)))

				rows := sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "created_at", "updated_at"})
				for _, u := range f.users {
					rows.AddRow(u.ID, u.Email, u.FirstName, u.LastName, u.CreatedAt, u.UpdatedAt)
				}
				mock.ExpectQuery(sqlListUsers).
					WithArgs(10, 0).
					WillReturnRows(rows)

				mock.ExpectCommit()
			},
			want:    f.users,
			wantN:   len(f.users),
			wantErr: nil,
		},
		{
			name:     "invalid page",
			page:     0,
			pageSize: 10,
			setup:    func(mock sqlmock.Sqlmock) {},
			want:     nil,
			wantN:    0,
			wantErr:  ErrInvalidPage,
		},
		{
			name:     "invalid page size",
			page:     1,
			pageSize: 0,
			setup:    func(mock sqlmock.Sqlmock) {},
			want:     nil,
			wantN:    0,
			wantErr:  ErrInvalidPageSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(f.mock)

			got, gotN, err := f.store.ListUsers(context.Background(), tt.page, tt.pageSize)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, got)
				assert.Zero(t, gotN)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.wantN, gotN)
			}

			assert.NoError(t, f.mock.ExpectationsWereMet(), "SQL expectations not met")
		})
	}
}

// Simplified test for UpdateUser - focusing on the key test cases
func TestUpdateUser(t *testing.T) {
	f := setupTest(t)
	defer f.cleanup()

	email := "updated@example.com"

	t.Run("invalid id", func(t *testing.T) {
		err := f.store.UpdateUser(context.Background(), 0, domain.UserUpdate{})
		assert.Equal(t, ErrInvalidID, err)
	})

	t.Run("user not found", func(t *testing.T) {
		f.mock.ExpectExec("UPDATE users SET email = $1, updated_at = NOW() WHERE id = $2 RETURNING id").
			WithArgs(email, int64(999)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := f.store.UpdateUser(context.Background(), 999, domain.UserUpdate{
			Email: email,
		})

		assert.Equal(t, ErrUserNotFound, err)
		assert.NoError(t, f.mock.ExpectationsWereMet())
	})
}
