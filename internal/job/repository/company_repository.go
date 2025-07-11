package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	commonerrors "github.com/benidevo/vega/internal/common/errors"
	"github.com/benidevo/vega/internal/job/models"
)

// SQLiteCompanyRepository is a SQLite implementation of CompanyRepository
type SQLiteCompanyRepository struct {
	db *sql.DB
}

// NewSQLiteCompanyRepository creates a new SQLiteCompanyRepository instance
func NewSQLiteCompanyRepository(db *sql.DB) *SQLiteCompanyRepository {
	return &SQLiteCompanyRepository{db: db}
}

// GetOrCreate retrieves a company by name or creates it if it doesn't exist
func (r *SQLiteCompanyRepository) GetOrCreate(ctx context.Context, userID int, name string) (*models.Company, error) {
	normalizedName, err := validateCompanyName(name)
	if err != nil {
		return nil, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, &commonerrors.RepositoryError{
			SentinelError: models.ErrTransactionFailed,
			InnerError:    err,
		}
	}

	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	var company models.Company
	err = tx.QueryRowContext(
		ctx,
		"SELECT id, name, created_at, updated_at FROM companies WHERE LOWER(name) = LOWER(?) AND user_id = ?",
		normalizedName, userID,
	).Scan(&company.ID, &company.Name, &company.CreatedAt, &company.UpdatedAt)

	if err == sql.ErrNoRows {
		now := time.Now().UTC()
		result, err := tx.ExecContext(
			ctx,
			"INSERT INTO companies (name, user_id, created_at, updated_at) VALUES (?, ?, ?, ?)",
			normalizedName, userID, now, now,
		)
		if err != nil {
			return nil, &commonerrors.RepositoryError{
				SentinelError: models.ErrFailedToCreateCompany,
				InnerError:    err,
			}
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, &commonerrors.RepositoryError{
				SentinelError: models.ErrFailedToCreateCompany,
				InnerError:    err,
			}
		}

		err = tx.Commit()
		if err != nil {
			return nil, &commonerrors.RepositoryError{
				SentinelError: models.ErrTransactionFailed,
				InnerError:    err,
			}
		}

		tx = nil

		company = models.Company{
			ID:        int(id),
			Name:      normalizedName,
			CreatedAt: now,
			UpdatedAt: now,
		}
		return &company, nil
	} else if err != nil {
		return nil, &commonerrors.RepositoryError{
			SentinelError: models.ErrCompanyNotFound,
			InnerError:    err,
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, &commonerrors.RepositoryError{
			SentinelError: models.ErrTransactionFailed,
			InnerError:    err,
		}
	}
	tx = nil

	return &company, nil
}

// wrapError is a helper function to create a repository error
func wrapError(sentinel, inner error) error {
	return &commonerrors.RepositoryError{
		SentinelError: sentinel,
		InnerError:    inner,
	}
}

// GetByID retrieves a company by its ID
func (r *SQLiteCompanyRepository) GetByID(ctx context.Context, userID int, id int) (*models.Company, error) {
	if id <= 0 {
		return nil, models.ErrInvalidCompanyID
	}

	query := "SELECT id, name, created_at, updated_at FROM companies WHERE id = ? AND user_id = ?"
	var company models.Company
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(&company.ID, &company.Name, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrCompanyNotFound
		}
		return nil, wrapError(models.ErrCompanyNotFound, err)
	}
	return &company, nil
}

// GetByName retrieves a company by its name
func (r *SQLiteCompanyRepository) GetByName(ctx context.Context, userID int, name string) (*models.Company, error) {
	normalizedName, err := validateCompanyName(name)
	if err != nil {
		return nil, err
	}

	query := "SELECT id, name, created_at, updated_at FROM companies WHERE LOWER(name) = LOWER(?) AND user_id = ?"
	var company models.Company
	err = r.db.QueryRowContext(ctx, query, normalizedName, userID).Scan(&company.ID, &company.Name, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrCompanyNotFound
		}
		return nil, wrapError(models.ErrCompanyNotFound, err)
	}

	return &company, nil
}

// validateCompanyName checks if the company name is valid and normalizes it
func validateCompanyName(name string) (string, error) {
	if name == "" {
		return "", models.ErrCompanyNameRequired
	}

	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		return "", models.ErrCompanyNameRequired
	}

	return normalizedName, nil
}

// GetAll retrieves all companies from the database
func (r *SQLiteCompanyRepository) GetAll(ctx context.Context, userID int) ([]*models.Company, error) {
	query := "SELECT id, name, created_at, updated_at FROM companies WHERE user_id = ? ORDER BY name"

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, wrapError(models.ErrFailedToCreateCompany, err)
	}
	defer rows.Close()

	var companies []*models.Company
	for rows.Next() {
		var company models.Company
		err := rows.Scan(&company.ID, &company.Name, &company.CreatedAt, &company.UpdatedAt)
		if err != nil {
			return nil, wrapError(models.ErrFailedToCreateCompany, err)
		}
		companies = append(companies, &company)
	}

	if err = rows.Err(); err != nil {
		return nil, wrapError(models.ErrFailedToCreateCompany, err)
	}

	return companies, nil
}

// Delete removes a company from the database by ID
func (r *SQLiteCompanyRepository) Delete(ctx context.Context, userID int, id int) error {
	if id <= 0 {
		return models.ErrInvalidCompanyID
	}

	query := "DELETE FROM companies WHERE id = ? AND user_id = ?"

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return wrapError(models.ErrFailedToDeleteCompany, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return wrapError(models.ErrFailedToDeleteCompany, err)
	}

	if rowsAffected == 0 {
		return models.ErrCompanyNotFound
	}

	return nil
}

// Update updates a company in the database
func (r *SQLiteCompanyRepository) Update(ctx context.Context, userID int, company *models.Company) error {
	if company == nil {
		return models.ErrInvalidCompanyID
	}

	if company.ID <= 0 {
		return models.ErrInvalidCompanyID
	}

	normalizedName, err := validateCompanyName(company.Name)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	query := "UPDATE companies SET name = ?, updated_at = ? WHERE id = ? AND user_id = ?"

	result, err := r.db.ExecContext(ctx, query, normalizedName, now, company.ID, userID)
	if err != nil {
		return wrapError(models.ErrFailedToUpdateCompany, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return wrapError(models.ErrFailedToUpdateCompany, err)
	}

	if rowsAffected == 0 {
		return models.ErrCompanyNotFound
	}

	company.Name = normalizedName
	company.UpdatedAt = now

	return nil
}
