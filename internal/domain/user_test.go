package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserCreateValidate(t *testing.T) {
	tests := []struct {
		name    string
		user    UserCreate
		wantErr bool
	}{
		{
			name: "valid user",
			user: UserCreate{
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: false,
		},
		{
			name: "empty email",
			user: UserCreate{
				Email:     "",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			user: UserCreate{
				Email:     "invalid-email",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: true,
		},
		{
			name: "empty first name",
			user: UserCreate{
				Email:     "test@example.com",
				FirstName: "",
				LastName:  "Doe",
			},
			wantErr: true,
		},
		{
			name: "empty last name",
			user: UserCreate{
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserUpdateValidate(t *testing.T) {
	validEmail := "test@example.com"
	invalidEmail := "not-an-email"
	validName := "John"

	tests := []struct {
		name    string
		update  UserUpdate
		wantErr bool
	}{
		{
			name: "valid update - all fields",
			update: UserUpdate{
				Email:     validEmail,
				FirstName: validName,
				LastName:  validName,
			},
			wantErr: false,
		},
		{
			name: "valid update - email only",
			update: UserUpdate{
				Email: validEmail,
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			update: UserUpdate{
				Email: invalidEmail,
			},
			wantErr: true,
		},
		{
			name:    "empty update",
			update:  UserUpdate{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.update.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_ToResponse(t *testing.T) {
	user := User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	response := user.ToResponse()

	assert.Equal(t, user.ID, response.ID)
	assert.Equal(t, user.Email, response.Email)
	assert.Equal(t, user.FirstName, response.FirstName)
	assert.Equal(t, user.LastName, response.LastName)
}
