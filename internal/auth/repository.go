package auth

import (
	"context"
	"database/sql"
	"time"
)

// UserRepository defines the interface for user-related data operations.
type UserRepository interface {
	// CreateUser inserts a new user into the database with the specified username,
	// password, and role, and returns the created User object.
	// It returns an error if the user creation fails.
	CreateUser(ctx context.Context, username, password, role string) (*User, error)

	// FindByUsername retrieves a user by their username.
	// It returns ErrUserNotFound if no user is found.
	FindByUsername(ctx context.Context, username string) (*User, error)

	// FindByID retrieves a user by their ID.
	// It returns ErrUserNotFound if no user is found.
	FindByID(ctx context.Context, id int) (*User, error)

	// UpdateUser updates the details of an existing user.
	// It returns ErrUserNotFound if the user does not exist.
	UpdateUser(ctx context.Context, user *User) (*User, error)

	// DeleteUser deletes a user by their ID.
	// It returns an error if the deletion fails.
	DeleteUser(ctx context.Context, id int) error

	// FindAllUsers retrieves all users from the database.
	// It returns an empty slice if no users are found.
	FindAllUsers(ctx context.Context) ([]*User, error)
}

// SQLiteUserRepository provides methods to interact with the user data
// stored in an SQLite database.
type SQLiteUserRepository struct {
	db *sql.DB
}

// NewSQLiteUserRepository creates a new instance of SQLiteUserRepository
// with the provided database connection.
func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

// CreateUser inserts a new user into the database with the specified username,
// password, and role, and returns the created User object.
// If a user with the given username already exists, it returns the existing user
// along with ErrUserAlreadyExists.
func (r *SQLiteUserRepository) CreateUser(ctx context.Context, username, password, role string) (*User, error) {
	existingUser, err := r.FindByUsername(ctx, username)
	if err == nil {
		return existingUser, ErrUserAlreadyExists
	} else if err != ErrUserNotFound {
		return nil, err
	}

	query := "INSERT INTO users (username, password, role) VALUES (?, ?, ?)"

	roleValue, err := RoleFromString(role)
	if err != nil {
		return nil, err
	}
	result, err := r.db.ExecContext(ctx, query, username, password, roleValue)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: users.username" {
			existingUser, findErr := r.FindByUsername(ctx, username)
			if findErr == nil {
				return existingUser, ErrUserAlreadyExists
			}
			return nil, ErrUserAlreadyExists
		}
		return nil, ErrUserCreationFailed
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, ErrUserCreationFailed
	}
	return r.FindByID(ctx, int(id))
}

// FindByID retrieves a user by their ID from the SQLite database.
func (r *SQLiteUserRepository) FindByID(ctx context.Context, id int) (*User, error) {
	query := "SELECT id, username, password, role, created_at, updated_at, last_login FROM users WHERE id = ?"

	var user User
	var lastLogin sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLogin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}

	return &user, nil
}

// FindByUsername retrieves a user from the database by their username.
func (r *SQLiteUserRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	query := "SELECT id, username, password, role, created_at, updated_at, last_login FROM users WHERE username = ?"

	var user User
	var lastLogin sql.NullTime
	err := r.db.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt, &lastLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}

	return &user, nil
}

// UpdateUser updates an existing user's details in the database and returns the updated user.
func (r *SQLiteUserRepository) UpdateUser(ctx context.Context, user *User) (*User, error) {
	query := "UPDATE users SET username = ?, password = ?, role = ?, updated_at = ? WHERE id = ?"

	user.UpdatedAt = time.Now().UTC()

	result, err := r.db.ExecContext(ctx, query, user.Username, user.Password, user.Role, user.UpdatedAt, user.ID)
	if err != nil {
		return nil, ErrUserUpdateFailed
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, ErrUserUpdateFailed
	}

	if rowsAffected == 0 {
		return nil, ErrUserNotFound
	}

	return r.FindByID(ctx, user.ID)
}

// DeleteUser removes a user from the database by their ID.
func (r *SQLiteUserRepository) DeleteUser(ctx context.Context, id int) error {
	query := "DELETE FROM users WHERE id = ?"

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return ErrUserDeletionFailed
	}
	return nil
}

// FindAllUsers retrieves all users from the database.
func (r *SQLiteUserRepository) FindAllUsers(ctx context.Context) ([]*User, error) {
	query := "SELECT id, username, password, role, created_at, updated_at, last_login FROM users"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, ErrUserListRetrievalFailed
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		var lastLogin sql.NullTime

		err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt, &lastLogin)
		if err != nil {
			return nil, ErrUserListRetrievalFailed
		}

		if lastLogin.Valid {
			user.LastLogin = lastLogin.Time
		}
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, ErrUserListRetrievalFailed
	}
	return users, nil
}
