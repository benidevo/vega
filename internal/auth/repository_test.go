package auth

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_login TIMESTAMP
		)
	`)
	require.NoError(t, err)

	return db
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)
	ctx := context.Background()

	t.Run("should create a new user", func(t *testing.T) {
		user, err := repo.CreateUser(ctx, "testuser", "password123", "admin")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "password123", user.Password)
		assert.Equal(t, ADMIN, user.Role)
		assert.NotZero(t, user.ID)
		assert.NotZero(t, user.CreatedAt)
		assert.NotZero(t, user.UpdatedAt)
		assert.Zero(t, user.LastLogin)
	})

	t.Run("should fail with duplicate username", func(t *testing.T) {
		user, err := repo.CreateUser(ctx, "testuser2", "anotherpassword", "standard")
		require.NoError(t, err)
		require.NotNil(t, user)

		existingUser, err := repo.CreateUser(ctx, "testuser2", "anotherpassword", "standard")
		assert.Error(t, err)
		assert.Equal(t, ErrUserAlreadyExists, err)
		assert.NotNil(t, existingUser) // The repository returns the existing user along with the error
		assert.Equal(t, user.ID, existingUser.ID)
	})
}

func TestFindUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)
	ctx := context.Background()

	createdUser, err := repo.CreateUser(ctx, "testuser", "password123", "admin")
	require.NoError(t, err)

	t.Run("should find a user by ID", func(t *testing.T) {
		user, err := repo.FindByID(ctx, createdUser.ID)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, createdUser.ID, user.ID)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "password123", user.Password)
		assert.Equal(t, ADMIN, user.Role)
	})

	t.Run("should find a user by username", func(t *testing.T) {
		user, err := repo.FindByUsername(ctx, "testuser")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, createdUser.ID, user.ID)
		assert.Equal(t, "testuser", user.Username)
	})

	t.Run("should return error for non-existent user ID", func(t *testing.T) {
		user, err := repo.FindByID(ctx, 9999)
		assert.Error(t, err)
		assert.Equal(t, ErrUserNotFound, err)
		assert.Nil(t, user)
	})

	t.Run("should return error for non-existent username", func(t *testing.T) {
		user, err := repo.FindByUsername(ctx, "nonexistentuser")
		assert.Error(t, err)
		assert.Equal(t, ErrUserNotFound, err)
		assert.Nil(t, user)
	})
}

func TestUpdateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)
	ctx := context.Background()

	createdUser, err := repo.CreateUser(ctx, "testuser", "password123", "admin")
	require.NoError(t, err)

	t.Run("should update user details", func(t *testing.T) {
		createdUser.Username = "updateduser"
		createdUser.Password = "newpassword"
		createdUser.Role = STANDARD
		createdUser.UpdatedAt = time.Now().UTC()

		updatedUser, err := repo.UpdateUser(ctx, createdUser)
		require.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, "updateduser", updatedUser.Username)
		assert.Equal(t, "newpassword", updatedUser.Password)
		assert.Equal(t, STANDARD, updatedUser.Role)

		// Verify the update was persisted
		fetchedUser, err := repo.FindByID(ctx, createdUser.ID)
		require.NoError(t, err)
		assert.Equal(t, "updateduser", fetchedUser.Username)
		assert.Equal(t, "newpassword", fetchedUser.Password)
		assert.Equal(t, STANDARD, fetchedUser.Role)
	})

	t.Run("should return error for non-existent user", func(t *testing.T) {
		nonExistentUser := &User{
			ID:        9999,
			Username:  "nonexistent",
			Password:  "password123",
			Role:      STANDARD,
			UpdatedAt: time.Now().UTC(),
		}

		updatedUser, err := repo.UpdateUser(ctx, nonExistentUser)
		assert.Error(t, err)
		assert.Nil(t, updatedUser)
	})
}

func TestDeleteUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)
	ctx := context.Background()

	// Create a test user
	createdUser, err := repo.CreateUser(ctx, "testuser", "password123", "admin")
	require.NoError(t, err)

	t.Run("should delete a user", func(t *testing.T) {
		err := repo.DeleteUser(ctx, createdUser.ID)
		require.NoError(t, err)

		// Verify the user was deleted
		user, err := repo.FindByID(ctx, createdUser.ID)
		assert.Error(t, err)
		assert.Equal(t, ErrUserNotFound, err)
		assert.Nil(t, user)
	})

	t.Run("should not return error when deleting non-existent user", func(t *testing.T) {
		err := repo.DeleteUser(ctx, 9999)
		assert.NoError(t, err)
	})
}

func TestFindAllUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)
	ctx := context.Background()

	t.Run("should return empty slice when no users exist", func(t *testing.T) {
		users, err := repo.FindAllUsers(ctx)
		require.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("should return all users", func(t *testing.T) {
		// Create test users
		user1, err := repo.CreateUser(ctx, "user1", "password1", "admin")
		require.NoError(t, err)

		user2, err := repo.CreateUser(ctx, "user2", "password2", "standard")
		require.NoError(t, err)

		users, err := repo.FindAllUsers(ctx)
		require.NoError(t, err)
		assert.Len(t, users, 2)

		// Check that both users are in the result
		var foundUser1, foundUser2 bool
		for _, u := range users {
			if u.ID == user1.ID {
				foundUser1 = true
				assert.Equal(t, "user1", u.Username)
				assert.Equal(t, ADMIN, u.Role)
			}
			if u.ID == user2.ID {
				foundUser2 = true
				assert.Equal(t, "user2", u.Username)
				assert.Equal(t, STANDARD, u.Role)
			}
		}
		assert.True(t, foundUser1, "user1 should be in the results")
		assert.True(t, foundUser2, "user2 should be in the results")
	})
}
