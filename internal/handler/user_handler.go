package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/huberts90/restful-api/internal/domain"
	"github.com/huberts90/restful-api/internal/storage"
	"go.uber.org/zap"
)

// UserHandler handles HTTP requests related to users
type UserHandler struct {
	store  storage.Storer
	logger *zap.Logger
}

// NewUserHandler creates a new UserHandler with the given dependencies
func NewUserHandler(store storage.Storer, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		store:  store,
		logger: logger,
	}
}

// RegisterRoutes registers all the user-related routes with the router
// This method centralizes route configuration, making it easier to understand the API
func (h *UserHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/users", h.CreateUser).Methods(http.MethodPost)
	router.HandleFunc("/users/{id:[0-9]+}", h.GetUser).Methods(http.MethodGet)
	router.HandleFunc("/users/{id:[0-9]+}", h.UpdateUser).Methods(http.MethodPut)
	router.HandleFunc("/users/{id:[0-9]+}", h.DeleteUser).Methods(http.MethodDelete)
	router.HandleFunc("/users", h.ListUsers).Methods(http.MethodGet)
}

// CreateUser handles the creation of a new user
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// Parse and validate the request body
	var userCreate domain.UserCreate
	// TODO: log request details
	if err := json.NewDecoder(r.Body).Decode(&userCreate); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if err := userCreate.Validate(); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create a context with timeout for the database operation
	// @MENTION_ME: Query the database with a "fuse"
	// TODO: parametrise timeout
	ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
	defer cancel()

	// Create the user
	userID, err := h.store.CreateUser(ctx, userCreate)
	if err != nil {
		if errors.Is(err, storage.ErrDuplicateEmail) {
			h.respondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		h.logger.Error("Failed to create user", zap.Error(err), zap.String("email", userCreate.Email))
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	h.respondWithData(w, http.StatusCreated, domain.UserCreateResponse{ID: userID})
}

// GetUser handles retrieving a user by ID
// Parses the ID from the URL, fetches the user, and returns it
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseIDFromURL(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Create a context with timeout for the database operation
	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
	defer cancel()

	// Get the user
	user, err := h.store.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			h.respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		h.logger.Error("Failed to get user", zap.Error(err), zap.Int64("id", id))
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	h.respondWithData(w, http.StatusOK, user.ToResponse())
}

// UpdateUser handles updating a user by ID
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseIDFromURL(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Parse and validate the request body
	var userUpdate domain.UserUpdate
	if err := json.NewDecoder(r.Body).Decode(&userUpdate); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := userUpdate.Validate(); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create a context with timeout for the database operation
	ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
	defer cancel()

	// Update the user
	err = h.store.UpdateUser(ctx, id, userUpdate)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			h.respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		if errors.Is(err, storage.ErrDuplicateEmail) {
			h.respondWithError(w, http.StatusConflict, "Email already exists")
			return
		}

		h.logger.Error("Failed to update user", zap.Error(err), zap.Int64("id", id))
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	h.respondWithData(w, http.StatusOK, "")
}

// DeleteUser handles deleting a user by ID
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseIDFromURL(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Create a context with timeout for the database operation
	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
	defer cancel()

	// Delete the user
	if err := h.store.DeleteUser(ctx, id); err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			h.respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		h.logger.Error("Failed to delete user", zap.Error(err), zap.Int64("id", id))
		h.respondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListUsers handles retrieving a paginated list of users
// Supports pagination via query parameters
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil || pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// Create a context with timeout for the database operation
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	// List users
	users, totalCount, err := h.store.ListUsers(ctx, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	// @MENTION_ME: direct access to the field in loop is more efficient
	usersToResponse := make([]domain.UserResponse, len(users))
	for i, usr := range users {
		usersToResponse[i] = usr.ToResponse()
	}

	// Calculate total pages (with minimum of 1)
	totalPages := totalCount / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	response := domain.PaginatedUsersResponse{
		Users:      usersToResponse,
		TotalCount: totalCount,
		TotalPages: totalPages,
		Page:       page,
		PageSize:   pageSize,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// Helper function to respond with JSON
func (h *UserHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		h.logger.Error("failed to marshal JSON response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err = w.Write(response); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}

// Helper function to respond with an error
// Standardizes error response format across the API
func (h *UserHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

// Helper function to respond with data
// Centralizes data response creation to avoid code duplication
func (h *UserHandler) respondWithData(w http.ResponseWriter, code int, payload interface{}) {
	h.respondWithJSON(w, code, payload)
}

// Helper function to extract and parse user ID from the URL
// Returns an error if the ID is invalid
func (h *UserHandler) parseIDFromURL(r *http.Request) (int64, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid user ID: %v", vars["id"])
	}
	return id, nil
}
