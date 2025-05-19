package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/benidevo/prospector/internal/auth/models"
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

func TestUserRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteUserRepository(db)
	ctx := context.Background()

	createdUser, err := repo.CreateUser(ctx, "testuser", "password123", "admin")
	require.NoError(t, err)
	assert.Equal(t, "testuser", createdUser.Username)
	assert.Equal(t, models.ADMIN, createdUser.Role)
	assert.NotZero(t, createdUser.ID)

	t.Run("should_return_error_when_creating_duplicate_username", func(t *testing.T) {
		user, err := repo.CreateUser(ctx, "testuser", "anotherpassword", "standard")
		assert.Error(t, err)
		assert.Equal(t, models.ErrUserAlreadyExists, err)
		assert.NotNil(t, user)
		assert.Equal(t, createdUser.ID, user.ID)
	})

	t.Run("should_create_new_user_when_username_is_unique", func(t *testing.T) {
		newUser, err := repo.CreateUser(ctx, "testuser2", "password", "standard")
		require.NoError(t, err)
		assert.NotNil(t, newUser)
		assert.Equal(t, "testuser2", newUser.Username)
		assert.Equal(t, models.STANDARD, newUser.Role)
		assert.NotEqual(t, createdUser.ID, newUser.ID)
	})

	t.Run("should_find_user_when_id_exists", func(t *testing.T) {
		user, err := repo.FindByID(ctx, createdUser.ID)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, createdUser.ID, user.ID)
		assert.Equal(t, "testuser", user.Username)
	})

	t.Run("should_return_error_when_id_does_not_exist", func(t *testing.T) {
		user, err := repo.FindByID(ctx, 9999)
		assert.Error(t, err)
		assert.Equal(t, models.ErrUserNotFound, err)
		assert.Nil(t, user)
	})

	t.Run("should_find_user_when_username_exists", func(t *testing.T) {
		user, err := repo.FindByUsername(ctx, "testuser")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, createdUser.ID, user.ID)
		assert.Equal(t, "testuser", user.Username)
	})

	t.Run("should_return_error_when_username_does_not_exist", func(t *testing.T) {
		user, err := repo.FindByUsername(ctx, "nonexistentuser")
		assert.Error(t, err)
		assert.Equal(t, models.ErrUserNotFound, err)
		assert.Nil(t, user)
	})

	t.Run("should_update_user_when_it_exists", func(t *testing.T) {
		createdUser.Username = "updateduser"
		createdUser.Password = "newpassword"
		createdUser.Role = models.STANDARD
		createdUser.UpdatedAt = time.Now().UTC()

		updatedUser, err := repo.UpdateUser(ctx, createdUser)
		require.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, "updateduser", updatedUser.Username)
		assert.Equal(t, "newpassword", updatedUser.Password)
		assert.Equal(t, models.STANDARD, updatedUser.Role)

		fetchedUser, err := repo.FindByID(ctx, createdUser.ID)
		require.NoError(t, err)
		assert.Equal(t, "updateduser", fetchedUser.Username)
	})

	t.Run("should_return_error_when_updating_nonexistent_user", func(t *testing.T) {
		nonExistentUser := &models.User{
			ID:        9999,
			Username:  "nonexistent",
			Password:  "password123",
			Role:      models.STANDARD,
			UpdatedAt: time.Now().UTC(),
		}

		updatedUser, err := repo.UpdateUser(ctx, nonExistentUser)
		assert.Error(t, err)
		assert.Nil(t, updatedUser)
	})

	t.Run("should_return_all_users_when_multiple_exist", func(t *testing.T) {
		user2, err := repo.CreateUser(ctx, "anotheruser", "password456", "standard")
		require.NoError(t, err)

		users, err := repo.FindAllUsers(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(users), 2)

		var foundUser1, foundUser2 bool
		for _, u := range users {
			if u.ID == createdUser.ID {
				foundUser1 = true
				assert.Equal(t, "updateduser", u.Username) // The username was updated in previous test
			}
			if u.ID == user2.ID {
				foundUser2 = true
				assert.Equal(t, "anotheruser", u.Username)
			}
		}
		assert.True(t, foundUser1, "updateduser should be in the results")
		assert.True(t, foundUser2, "anotheruser should be in the results")
	})

	t.Run("should_delete_user_when_it_exists", func(t *testing.T) {
		err := repo.DeleteUser(ctx, createdUser.ID)
		require.NoError(t, err)

		deletedUser, err := repo.FindByID(ctx, createdUser.ID)
		assert.Error(t, err)
		assert.Equal(t, models.ErrUserNotFound, err)
		assert.Nil(t, deletedUser)
	})

	t.Run("should_not_error_when_deleting_nonexistent_user", func(t *testing.T) {
		err = repo.DeleteUser(ctx, 9999)
		assert.NoError(t, err)
	})
}
