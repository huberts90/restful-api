package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/huberts90/restful-api/internal/logger"
	storagemocks "github.com/huberts90/restful-api/internal/storage/mocks"

	"github.com/gorilla/mux"
	"github.com/huberts90/restful-api/internal/domain"
	"github.com/huberts90/restful-api/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// FIXME: add more test cases

func TestCreateUser(t *testing.T) {
	// Set up the mock store
	mockStore := storagemocks.NewMockStorer(t)
	handler := NewUserHandler(mockStore, logger.NewNoOpLogger())

	// Create a test user
	userID := int64(1)
	userCreate := domain.UserCreate{
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	// Mock the store's CreateUser method
	mockStore.On("CreateUser", mock.Anything, userCreate).Return(userID, nil)

	// Create a test request
	userJSON, _ := json.Marshal(userCreate)
	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(userJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler
	handler.CreateUser(rr, req)

	// Check the response
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Parse the response body
	var responseUser domain.UserCreateResponse
	err = json.Unmarshal(rr.Body.Bytes(), &responseUser)
	require.NoError(t, err)

	// Verify the response
	assert.Equal(t, userID, responseUser.ID)

	// Verify the store was called
	mockStore.AssertExpectations(t)
}

func TestGetUser(t *testing.T) {
	// Set up the mock store
	mockStore := storagemocks.NewMockStorer(t)
	handler := NewUserHandler(mockStore, logger.NewNoOpLogger())

	// Create a test user
	user := &domain.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock the store's GetUserByID method
	mockStore.On("GetUserByID", mock.Anything, int64(1)).Return(user, nil)

	// Create a test request
	req, err := http.NewRequest("GET", "/users/1", nil)
	require.NoError(t, err)

	// Add the ID parameter using gorilla/mux
	vars := map[string]string{
		"id": "1",
	}
	req = mux.SetURLVars(req, vars)

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler
	handler.GetUser(rr, req)

	// Check the response
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse the response body
	var responseUser domain.UserResponse
	err = json.Unmarshal(rr.Body.Bytes(), &responseUser)
	require.NoError(t, err)

	// Verify the response
	assert.Equal(t, user.ID, responseUser.ID)
	assert.Equal(t, user.Email, responseUser.Email)
	assert.Equal(t, user.FirstName, responseUser.FirstName)
	assert.Equal(t, user.LastName, responseUser.LastName)

	// Verify the store was called
	mockStore.AssertExpectations(t)
}

func TestGetUser_NotFound(t *testing.T) {
	// Set up the mock store
	mockStore := storagemocks.NewMockStorer(t)
	handler := NewUserHandler(mockStore, logger.NewNoOpLogger())

	// Mock the store's GetUserByID method to return a not found error
	mockStore.On("GetUserByID", mock.Anything, int64(999)).Return(nil, storage.ErrUserNotFound)

	// Create a test request
	req, err := http.NewRequest("GET", "/users/999", nil)
	require.NoError(t, err)

	// Add the ID parameter using gorilla/mux
	vars := map[string]string{
		"id": "999",
	}
	req = mux.SetURLVars(req, vars)

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler
	handler.GetUser(rr, req)

	// Check the response
	assert.Equal(t, http.StatusNotFound, rr.Code)

	// Verify the store was called
	mockStore.AssertExpectations(t)
}

func TestUpdateUser(t *testing.T) {
	// Set up the mock store
	mockStore := storagemocks.NewMockStorer(t)
	handler := NewUserHandler(mockStore, logger.NewNoOpLogger())

	// Create a test user update
	firstName := "Updated"
	lastName := "Name"
	userUpdate := domain.UserUpdate{
		FirstName: firstName,
		LastName:  lastName,
	}

	// Mock the store's UpdateUser method
	mockStore.On("UpdateUser", mock.Anything, int64(1), userUpdate).Return(nil)

	// Create a test request
	userJSON, _ := json.Marshal(userUpdate)
	req, err := http.NewRequest("PUT", "/users/1", bytes.NewBuffer(userJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Add the ID parameter using gorilla/mux
	vars := map[string]string{
		"id": "1",
	}
	req = mux.SetURLVars(req, vars)

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler
	handler.UpdateUser(rr, req)

	// Check the response
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify the store was called
	mockStore.AssertExpectations(t)
}

func TestDeleteUser(t *testing.T) {
	// Set up the mock store
	mockStore := storagemocks.NewMockStorer(t)
	handler := NewUserHandler(mockStore, logger.NewNoOpLogger())

	// Mock the store's UpdateUser method
	mockStore.On("DeleteUser", mock.Anything, int64(1)).Return(nil)

	// Create a test request
	req, err := http.NewRequest("DELETE", "/users/1", nil)
	require.NoError(t, err)

	// Add the ID parameter using gorilla/mux
	vars := map[string]string{
		"id": "1",
	}
	req = mux.SetURLVars(req, vars)

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler
	handler.DeleteUser(rr, req)

	// Check the status code - should be 204 No Content
	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Verify the store was called
	mockStore.AssertExpectations(t)
}

func TestListUser(t *testing.T) {
	// Set up the mock store
	mockStore := storagemocks.NewMockStorer(t)
	handler := NewUserHandler(mockStore, logger.NewNoOpLogger())

	now := time.Now()
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
	// Mock the store's UpdateUser method
	mockStore.On("ListUsers", mock.Anything, 2, 5).Return(users, 20, nil)

	// Create a test request
	req, err := http.NewRequest("GET", "/users?page=2&page_size=5", nil)
	require.NoError(t, err)

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ListUsers(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify the store was called
	mockStore.AssertExpectations(t)

	// Parse response
	var response domain.PaginatedUsersResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check meta values
	assert.Equal(t, 20, response.TotalCount)
	assert.Equal(t, 2, response.Page)
	assert.Equal(t, 5, response.PageSize)
	assert.Equal(t, 4, response.TotalPages)

	// Check length of data array
	assert.Equal(t, 2, len(response.Users))

	// Check first user if available
	assert.Equal(t, int64(1), response.Users[0].ID)
}
