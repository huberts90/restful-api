package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/huberts90/restful-api/internal/domain"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrDuplicateEmail   = errors.New("email already exists")
	ErrInvalidID        = errors.New("invalid ID")
	ErrInvalidPageSize  = errors.New("invalid page size")
	ErrInvalidPage      = errors.New("invalid page")
	ErrDatabaseInternal = errors.New("internal database error")
)

// SQL queries - ensure they have no extra whitespace for exact matching in tests
const (
	sqlCreateUser  = `INSERT INTO users (email, first_name, last_name, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`
	sqlGetUserByID = `SELECT id, email, first_name, last_name, created_at, updated_at FROM users WHERE id = $1`
	sqlDeleteUser  = `DELETE FROM users WHERE id = $1`
	sqlCountUsers  = `SELECT COUNT(*) FROM users`
	sqlListUsers   = `SELECT id, email, first_name, last_name, created_at, updated_at FROM users ORDER BY id LIMIT $1 OFFSET $2`
)

// PostgresStore implements the Storer interface using PostgreSQL
type PostgresStore struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPostgresStore creates a new PostgreSQL store
func NewPostgresStore(cfg PostgresConfig, logger *zap.Logger) (*PostgresStore, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		// @MENTION_ME: do not ignore errors
		defer func() {
			if closeErr := db.Close(); closeErr != nil {
				logger.Warn("failed to close database connection", zap.Error(closeErr))
			}
		}()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStore{
		db:     db,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// handleError is a helper method to handle database errors
func (s *PostgresStore) handleError(err error, msg string, fields ...zap.Field) error {
	// Check for duplicate email error
	const pgDuplicateCode = "23505"
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == pgDuplicateCode && pqErr.Constraint == "users_email_key" {
		return ErrDuplicateEmail
	}

	// Check for no rows error
	if errors.Is(err, sql.ErrNoRows) {
		return ErrUserNotFound
	}

	// Internal error
	errorFields := append([]zap.Field{zap.Error(err)}, fields...)
	s.logger.Error(msg, errorFields...)

	// @MENTION_ME: error wrapping
	return fmt.Errorf("%w: %w", ErrDatabaseInternal, err)
}

// withTx executes a function within a transaction
func (s *PostgresStore) withTx(ctx context.Context, readOnly bool, fn func(*sql.Tx) error) error {
	// @MENTION_ME:
	// - wrap operation within transaction
	// - always pass context
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: readOnly})
	if err != nil {
		return s.handleError(err, "failed to begin transaction")
	}
	defer func() {
		// @MENTION_ME: transaction in idle
		if txErr := tx.Rollback(); txErr != nil {
			s.logger.Error("failed to rollback transaction", zap.Error(txErr))
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return s.handleError(err, "failed to commit transaction")
	}

	return nil
}

// scanUser scans a user from a row
func (s *PostgresStore) scanUser(row interface{ Scan(...interface{}) error }) (*domain.User, error) {
	user := &domain.User{}
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// CreateUser inserts a new user into the database
func (s *PostgresStore) CreateUser(ctx context.Context, userCreate domain.UserCreate) (int64, error) {
	// @MENTION_ME: pq does not support the LastInsertId() method of the Result type in database/sql. To return the identifier of an INSERT (or UPDATE or DELETE), use the Postgres RETURNING clause with a standard Query or QueryRow call
	row := s.db.QueryRowContext(
		ctx,
		sqlCreateUser,
		userCreate.Email,
		userCreate.FirstName,
		userCreate.LastName,
	)
	var userID int64
	err := row.Scan(&userID)

	if err != nil {
		return 0, s.handleError(err, "failed to retrieve last inserted id")
	}

	return userID, nil
}

// GetUserByID retrieves a user by their ID
func (s *PostgresStore) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	row := s.db.QueryRowContext(ctx, sqlGetUserByID, id)
	user, err := s.scanUser(row)
	if err != nil {
		return nil, s.handleError(err, "failed to get user by ID", zap.Int64("id", id))
	}

	return user, nil
}

// buildUpdateQuery constructs the UPDATE query and arguments
func (s *PostgresStore) buildUpdateQuery(userUpdate domain.UserUpdate, id int64) (string, []interface{}) {
	var updates []string
	args := make([]interface{}, 0, 4) // @MENTION_ME: len + cap
	argPosition := 1

	if userUpdate.Email != "" {
		updates = append(updates, fmt.Sprintf("email = $%d", argPosition))
		args = append(args, userUpdate.Email)
		argPosition++
	}
	if userUpdate.FirstName != "" {
		updates = append(updates, fmt.Sprintf("first_name = $%d", argPosition))
		args = append(args, userUpdate.FirstName)
		argPosition++
	}
	if userUpdate.LastName != "" {
		updates = append(updates, fmt.Sprintf("last_name = $%d", argPosition))
		args = append(args, userUpdate.LastName)
		argPosition++
	}

	updates = append(updates, "updated_at = NOW()")
	query := "UPDATE users SET " + strings.Join(updates, ", ") +
		fmt.Sprintf(" WHERE id = $%d RETURNING id", argPosition)
	args = append(args, id)

	return query, args
}

// UpdateUser updates an existing user
func (s *PostgresStore) UpdateUser(ctx context.Context, id int64, userUpdate domain.UserUpdate) error {
	if id <= 0 {
		return ErrInvalidID
	}

	// Build and execute update query
	query, args := s.buildUpdateQuery(userUpdate, id)
	result, err := s.db.ExecContext(ctx, query, args...)
	zapFieldID := zap.Int64("id", id)
	if err != nil {
		return s.handleError(err, "failed to update user", zapFieldID)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return s.handleError(err, "failed to retrieve rows affected", zapFieldID)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeleteUser removes a user from the database
func (s *PostgresStore) DeleteUser(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidID
	}

	result, err := s.db.ExecContext(ctx, sqlDeleteUser, id)
	if err != nil {
		return s.handleError(err, "failed to delete user", zap.Int64("id", id))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return s.handleError(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ListUsers retrieves a paginated list of users
func (s *PostgresStore) ListUsers(ctx context.Context, page, pageSize int) ([]domain.User, int, error) {
	if page < 1 {
		return nil, 0, ErrInvalidPage
	}
	if pageSize < 1 || pageSize > 100 {
		return nil, 0, ErrInvalidPageSize
	}

	var users []domain.User
	var totalCount int

	err := s.withTx(ctx, true, func(tx *sql.Tx) error {
		// Get total count
		err := tx.QueryRowContext(ctx, sqlCountUsers).Scan(&totalCount)
		if err != nil {
			return s.handleError(err, "failed to count users")
		}

		// Get users for the current page
		offset := (page - 1) * pageSize
		rows, err := tx.QueryContext(ctx, sqlListUsers, pageSize, offset)
		if err != nil {
			return s.handleError(err, "failed to query users")
		}
		defer rows.Close()

		users = make([]domain.User, 0, pageSize)
		for rows.Next() {
			user, err := s.scanUser(rows)
			if err != nil {
				return s.handleError(err, "failed to scan user row")
			}
			users = append(users, *user)
		}

		if err = rows.Err(); err != nil {
			return s.handleError(err, "error iterating user rows")
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return users, totalCount, nil
}
