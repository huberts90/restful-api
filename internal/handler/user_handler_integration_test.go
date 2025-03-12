package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"github.com/huberts90/restful-api/internal/domain"
	"github.com/huberts90/restful-api/internal/handler"
	"github.com/huberts90/restful-api/internal/storage"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

// UserHandlerIntegrationSuite tests the user handler with a real database
// This test connects to a PostgreSQL instance provided by docker-compose
type UserHandlerIntegrationSuite struct {
	suite.Suite
	pgConfig storage.PostgresConfig
	store    *storage.PostgresStore
	handler  *handler.UserHandler
	router   *mux.Router
	logger   *zap.Logger
	db       *sql.DB
}

func (s *UserHandlerIntegrationSuite) SetupSuite() {
	t := s.T()

	// Create a development logger
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	s.logger = logger

	// Configure database connection from environment or use docker-compose defaults
	s.pgConfig = storage.PostgresConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		DBName:          "users_db",
		SSLMode:         "disable",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	// Create database store
	store, err := storage.NewPostgresStore(s.pgConfig, s.logger)
	require.NoError(t, err)
	s.store = store

	// Create a direct database connection for migrations
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		s.pgConfig.Host, s.pgConfig.Port, s.pgConfig.User, s.pgConfig.Password, s.pgConfig.DBName, s.pgConfig.SSLMode,
	)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	s.db = db
	s.db.Exec("TRUNCATE TABLE users")

	// Set up handler and router
	s.handler = handler.NewUserHandler(store, s.logger)

	router := mux.NewRouter()
	s.router = router.PathPrefix("/api").Subrouter()
	s.handler.RegisterRoutes(s.router)
}

func (s *UserHandlerIntegrationSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.store != nil {
		s.store.Close()
	}
}

// TestCreateUser tests creating a new user
func (s *UserHandlerIntegrationSuite) TestCreateUser() {
	t := s.T()

	// Test data
	userCreate := domain.UserCreate{
		Email:     "new@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	// Convert to JSON
	jsonData, err := json.Marshal(userCreate)
	require.NoError(t, err)

	// Create a request
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	// Serve the request
	s.router.ServeHTTP(recorder, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, recorder.Code)

	// Parse response
	var response domain.UserCreateResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify user was created with ID
	fmt.Println(response.ID)
	assert.NotZero(t, response.ID)

	// Test duplicate email
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	s.router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusConflict, recorder.Code)
}

// Run the test suite
func TestUserHandlerIntegrationSuite(t *testing.T) {
	// Skip if explicitly disabled
	if os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("Skipping integration tests")
	}

	// Run the tests
	suite.Run(t, new(UserHandlerIntegrationSuite))
}
