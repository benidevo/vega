package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/benidevo/ascentio/internal/job/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupMockDB creates a new mock database for testing
func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return db, mock
}

func TestSQLiteCompanyRepository_GetOrCreate(t *testing.T) {
	t.Run("should create company when it doesn't exist", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		repo := NewSQLiteCompanyRepository(db)
		companyName := "Test Company"
		normalizedName := companyName

		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"})
		mock.ExpectQuery("SELECT id, name, created_at, updated_at FROM companies WHERE LOWER\\(name\\) = LOWER\\(\\?\\)").
			WithArgs(normalizedName).
			WillReturnRows(rows)

		mock.ExpectExec("INSERT INTO companies \\(name, created_at, updated_at\\) VALUES \\(\\?, \\?, \\?\\)").
			WithArgs(normalizedName, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1)) // ID 1, 1 row affected

		mock.ExpectCommit()

		ctx := context.Background()
		company, err := repo.GetOrCreate(ctx, companyName)

		require.NoError(t, err)
		require.NotNil(t, company)
		assert.Equal(t, 1, company.ID)
		assert.Equal(t, normalizedName, company.Name)
		assert.NotZero(t, company.CreatedAt)
		assert.NotZero(t, company.UpdatedAt)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("should return existing company when it exists", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		repo := NewSQLiteCompanyRepository(db)
		companyName := "Test Company"
		normalizedName := companyName
		companyID := 2

		createdAt := time.Now().Add(-24 * time.Hour) // yesterday
		updatedAt := time.Now().Add(-1 * time.Hour)  // 1 hour ago

		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
			AddRow(companyID, normalizedName, createdAt, updatedAt)

		mock.ExpectQuery("SELECT id, name, created_at, updated_at FROM companies WHERE LOWER\\(name\\) = LOWER\\(\\?\\)").
			WithArgs(normalizedName).
			WillReturnRows(rows)

		mock.ExpectCommit()

		ctx := context.Background()
		company, err := repo.GetOrCreate(ctx, companyName)

		require.NoError(t, err)
		require.NotNil(t, company)
		assert.Equal(t, companyID, company.ID)
		assert.Equal(t, normalizedName, company.Name)
		assert.Equal(t, createdAt.Unix(), company.CreatedAt.Unix()) // Compare Unix timestamps due to serialization differences
		assert.Equal(t, updatedAt.Unix(), company.UpdatedAt.Unix())

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestSQLiteCompanyRepository_GetByID(t *testing.T) {
	t.Run("should return company when it exists", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		repo := NewSQLiteCompanyRepository(db)
		companyID := 1
		companyName := "Test Company"
		createdAt := time.Now().Add(-24 * time.Hour)
		updatedAt := time.Now()

		rows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
			AddRow(companyID, companyName, createdAt, updatedAt)

		mock.ExpectQuery("SELECT id, name, created_at, updated_at FROM companies WHERE id = ?").
			WithArgs(companyID).
			WillReturnRows(rows)

		ctx := context.Background()
		company, err := repo.GetByID(ctx, companyID)

		require.NoError(t, err)
		require.NotNil(t, company)
		assert.Equal(t, companyID, company.ID)
		assert.Equal(t, companyName, company.Name)
		assert.Equal(t, createdAt.Unix(), company.CreatedAt.Unix())
		assert.Equal(t, updatedAt.Unix(), company.UpdatedAt.Unix())

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("should return ErrCompanyNotFound when company does not exist", func(t *testing.T) {
		// Set up mock db
		db, mock := setupMockDB(t)
		defer db.Close()

		repo := NewSQLiteCompanyRepository(db)
		companyID := 999 // Non-existent company

		mock.ExpectQuery("SELECT id, name, created_at, updated_at FROM companies WHERE id = ?").
			WithArgs(companyID).
			WillReturnError(sql.ErrNoRows)

		ctx := context.Background()
		company, err := repo.GetByID(ctx, companyID)

		assert.Error(t, err)
		assert.Equal(t, ErrCompanyNotFound, err)
		assert.Nil(t, company)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestSQLiteCompanyRepository_GetByName(t *testing.T) {
	t.Run("should return company when it exists", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		repo := NewSQLiteCompanyRepository(db)
		companyID := 1
		companyName := "Test Company"
		createdAt := time.Now().Add(-24 * time.Hour)
		updatedAt := time.Now()

		rows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
			AddRow(companyID, companyName, createdAt, updatedAt)

		mock.ExpectQuery("SELECT id, name, created_at, updated_at FROM companies WHERE LOWER\\(name\\) = LOWER\\(\\?\\)").
			WithArgs(companyName).
			WillReturnRows(rows)

		ctx := context.Background()
		company, err := repo.GetByName(ctx, companyName)

		require.NoError(t, err)
		require.NotNil(t, company)
		assert.Equal(t, companyID, company.ID)
		assert.Equal(t, companyName, company.Name)
		assert.Equal(t, createdAt.Unix(), company.CreatedAt.Unix())
		assert.Equal(t, updatedAt.Unix(), company.UpdatedAt.Unix())

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("should return error for empty company name", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		repo := NewSQLiteCompanyRepository(db)

		ctx := context.Background()
		company, err := repo.GetByName(ctx, "")

		assert.Error(t, err)
		assert.Equal(t, ErrCompanyNameRequired, err)
		assert.Nil(t, company)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestSQLiteCompanyRepository_GetAll(t *testing.T) {
	t.Run("should return all companies", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		repo := NewSQLiteCompanyRepository(db)
		testTime := time.Now()

		companies := []struct {
			id        int
			name      string
			createdAt time.Time
			updatedAt time.Time
		}{
			{1, "Company A", testTime.Add(-48 * time.Hour), testTime.Add(-24 * time.Hour)},
			{2, "Company B", testTime.Add(-24 * time.Hour), testTime.Add(-12 * time.Hour)},
			{3, "Company C", testTime.Add(-12 * time.Hour), testTime},
		}

		rows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"})
		for _, c := range companies {
			rows.AddRow(c.id, c.name, c.createdAt, c.updatedAt)
		}

		mock.ExpectQuery("SELECT id, name, created_at, updated_at FROM companies ORDER BY name").
			WillReturnRows(rows)

		ctx := context.Background()
		result, err := repo.GetAll(ctx)

		require.NoError(t, err)
		require.Len(t, result, len(companies))

		for i, c := range companies {
			assert.Equal(t, c.id, result[i].ID)
			assert.Equal(t, c.name, result[i].Name)
			assert.Equal(t, c.createdAt.Unix(), result[i].CreatedAt.Unix())
			assert.Equal(t, c.updatedAt.Unix(), result[i].UpdatedAt.Unix())
		}

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestSQLiteCompanyRepository_Delete(t *testing.T) {
	t.Run("should delete company when it exists", func(t *testing.T) {
		// Set up mock db
		db, mock := setupMockDB(t)
		defer db.Close()

		repo := NewSQLiteCompanyRepository(db)
		companyID := 1

		mock.ExpectExec("DELETE FROM companies WHERE id = ?").
			WithArgs(companyID).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 0 last insert id, 1 row affected

		ctx := context.Background()
		err := repo.Delete(ctx, companyID)

		assert.NoError(t, err)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestSQLiteCompanyRepository_Update(t *testing.T) {
	t.Run("should update company when it exists", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		repo := NewSQLiteCompanyRepository(db)
		company := &models.Company{
			ID:   1,
			Name: "Updated Company Name",
		}

		mock.ExpectExec("UPDATE companies SET name = \\?, updated_at = \\? WHERE id = \\?").
			WithArgs(company.Name, sqlmock.AnyArg(), company.ID).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 0 last insert id, 1 row affected

		ctx := context.Background()
		err := repo.Update(ctx, company)

		assert.NoError(t, err)
		assert.NotZero(t, company.UpdatedAt, "Expected UpdatedAt to be set")

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}
