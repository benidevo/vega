package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/benidevo/vega/internal/job/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testUserID = 1

// setupMockDB creates a new mock database for testing
func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return db, mock
}

func TestSQLiteCompanyRepository_GetOrCreate(t *testing.T) {
	tests := []struct {
		name         string
		companyName  string
		existingRows []struct {
			id        int
			name      string
			createdAt time.Time
			updatedAt time.Time
		}
		expectInsert bool
		wantErr      bool
		validateFunc func(*testing.T, *models.Company)
	}{
		{
			name:         "creates new company",
			companyName:  "New Company",
			existingRows: nil,
			expectInsert: true,
			validateFunc: func(t *testing.T, c *models.Company) {
				assert.Equal(t, 1, c.ID)
				assert.Equal(t, "New Company", c.Name)
				assert.NotZero(t, c.CreatedAt)
				assert.NotZero(t, c.UpdatedAt)
			},
		},
		{
			name:        "returns existing company",
			companyName: "Existing Company",
			existingRows: []struct {
				id        int
				name      string
				createdAt time.Time
				updatedAt time.Time
			}{
				{2, "Existing Company", time.Now().Add(-24 * time.Hour), time.Now().Add(-1 * time.Hour)},
			},
			expectInsert: false,
			validateFunc: func(t *testing.T, c *models.Company) {
				assert.Equal(t, 2, c.ID)
				assert.Equal(t, "Existing Company", c.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			repo := NewSQLiteCompanyRepository(db)

			mock.ExpectBegin()

			// Setup SELECT expectation
			rows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"})
			for _, r := range tt.existingRows {
				rows.AddRow(r.id, r.name, r.createdAt, r.updatedAt)
			}

			mock.ExpectQuery("SELECT id, name, created_at, updated_at FROM companies WHERE LOWER\\(name\\) = LOWER\\(\\?\\) AND user_id = \\?").
				WithArgs(tt.companyName, testUserID).
				WillReturnRows(rows)

			if tt.expectInsert {
				mock.ExpectExec("INSERT INTO companies \\(name, user_id, created_at, updated_at\\) VALUES \\(\\?, \\?, \\?, \\?\\)").
					WithArgs(tt.companyName, testUserID, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			}

			mock.ExpectCommit()

			company, err := repo.GetOrCreate(context.Background(), testUserID, tt.companyName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, company)
				if tt.validateFunc != nil {
					tt.validateFunc(t, company)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSQLiteCompanyRepository_GetByID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSQLiteCompanyRepository(db)
	ctx := context.Background()

	t.Run("existing company", func(t *testing.T) {
		companyID := 1
		rows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
			AddRow(companyID, "Test Company", time.Now(), time.Now())

		mock.ExpectQuery("SELECT id, name, created_at, updated_at FROM companies WHERE id = \\? AND user_id = \\?").
			WithArgs(companyID, testUserID).
			WillReturnRows(rows)

		company, err := repo.GetByID(ctx, testUserID, companyID)

		require.NoError(t, err)
		require.NotNil(t, company)
		assert.Equal(t, companyID, company.ID)
		assert.Equal(t, "Test Company", company.Name)
	})

	t.Run("non-existent company", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, name, created_at, updated_at FROM companies WHERE id = \\? AND user_id = \\?").
			WithArgs(999, testUserID).
			WillReturnError(sql.ErrNoRows)

		company, err := repo.GetByID(ctx, testUserID, 999)

		assert.Error(t, err)
		assert.Equal(t, models.ErrCompanyNotFound, err)
		assert.Nil(t, company)
	})
}

func TestSQLiteCompanyRepository_GetByName(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSQLiteCompanyRepository(db)
	ctx := context.Background()

	t.Run("empty name returns error", func(t *testing.T) {
		company, err := repo.GetByName(ctx, testUserID, "")

		assert.Error(t, err)
		assert.Equal(t, models.ErrCompanyNameRequired, err)
		assert.Nil(t, company)
	})

	t.Run("existing company", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
			AddRow(1, "Test Company", time.Now(), time.Now())

		mock.ExpectQuery("SELECT id, name, created_at, updated_at FROM companies WHERE LOWER\\(name\\) = LOWER\\(\\?\\) AND user_id = \\?").
			WithArgs("Test Company", testUserID).
			WillReturnRows(rows)

		company, err := repo.GetByName(ctx, testUserID, "Test Company")

		require.NoError(t, err)
		require.NotNil(t, company)
		assert.Equal(t, "Test Company", company.Name)
	})
}

func TestSQLiteCompanyRepository_GetAll(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSQLiteCompanyRepository(db)
	testTime := time.Now()

	rows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
		AddRow(1, "Company A", testTime, testTime).
		AddRow(2, "Company B", testTime, testTime).
		AddRow(3, "Company C", testTime, testTime)

	mock.ExpectQuery("SELECT id, name, created_at, updated_at FROM companies WHERE user_id = \\? ORDER BY name").
		WithArgs(testUserID).
		WillReturnRows(rows)

	companies, err := repo.GetAll(context.Background(), testUserID)

	require.NoError(t, err)
	require.Len(t, companies, 3)
	assert.Equal(t, "Company A", companies[0].Name)
	assert.Equal(t, "Company B", companies[1].Name)
	assert.Equal(t, "Company C", companies[2].Name)
}

func TestSQLiteCompanyRepository_Delete(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSQLiteCompanyRepository(db)

	mock.ExpectExec("DELETE FROM companies WHERE id = \\? AND user_id = \\?").
		WithArgs(1, testUserID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), testUserID, 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLiteCompanyRepository_Update(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewSQLiteCompanyRepository(db)
	company := &models.Company{
		ID:   1,
		Name: "Updated Company",
	}

	mock.ExpectExec("UPDATE companies SET name = \\?, updated_at = \\? WHERE id = \\? AND user_id = \\?").
		WithArgs(company.Name, sqlmock.AnyArg(), company.ID, testUserID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), testUserID, company)

	assert.NoError(t, err)
	assert.NotZero(t, company.UpdatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}
