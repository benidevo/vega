package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	user, err := NewUser("testuser", "password123", ADMIN)

	require.NoError(t, err, "Expected no error when creating user")
	require.NotNil(t, user, "Expected user to be initialized")
	assert.Equal(t, "testuser", user.Username, "Expected username to be 'testuser'")
	assert.Equal(t, "password123", user.Password, "Expected password to be 'password123'")
	assert.Equal(t, ADMIN, user.Role, "Expected role to be ADMIN")
	assert.NotZero(t, user.CreatedAt, "Expected CreatedAt to be set")
	assert.NotZero(t, user.UpdatedAt, "Expected UpdatedAt to be set")
	assert.Zero(t, user.LastLogin, "Expected LastLogin to be zero value")
}

func TestNewUserValidation(t *testing.T) {
	t.Run("should return error for empty username", func(t *testing.T) {
		user, err := NewUser("", "password123", STANDARD)
		require.Error(t, err, "Expected error when username is empty")
		require.Nil(t, user, "Expected user to be nil")
		assert.Equal(t, "username cannot be empty", err.Error(), "Expected error message to match")
	})

	t.Run("should return error for empty password", func(t *testing.T) {
		user, err := NewUser("testuser", "", STANDARD)
		require.Error(t, err, "Expected error when password is empty")
		require.Nil(t, user, "Expected user to be nil")
		assert.Equal(t, "password cannot be empty", err.Error(), "Expected error message to match")
	})

}

func TestRole(t *testing.T) {
	t.Run("should return correct string for ADMIN role", func(t *testing.T) {
		role := ADMIN
		actual, err := role.String()
		require.NoError(t, err, "Expected no error when getting role string")
		assert.Equal(t, "Admin", actual, "Expected role string to be 'Admin'")
	})

	t.Run("should return correct string for STANDARD role", func(t *testing.T) {
		role := STANDARD
		actual, err := role.String()
		require.NoError(t, err, "Expected no error when getting role string")
		assert.Equal(t, "Standard", actual, "Expected role string to be 'Standard'")
	})

	t.Run("should return 'Unknown' for unknown role", func(t *testing.T) {
		role := Role(999)
		_, err := role.String()
		require.Error(t, err, "Expected error when getting string for unknown role")
	})

	t.Run("should return correct role for valid strings", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected Role
		}{
			{"admin", ADMIN},
			{"Standard", STANDARD},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				actual, err := RoleFromString(tc.input)
				require.NoError(t, err, "Expected no error when getting role from string")
				assert.Equal(t, tc.expected, actual, "Expected role to match expected value")
			})
		}
	})

	t.Run("should return error for invalid role string", func(t *testing.T) {
		role, err := RoleFromString("invalid")
		require.Error(t, err, "Expected error when getting role from invalid string")
		assert.Equal(t, -1, int(role), "Expected role to be -1 for invalid role")
	})
}
