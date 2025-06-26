package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/benidevo/vega/internal/job/models"
	"github.com/stretchr/testify/assert"
)

func TestSQLiteJobRepository_MatchResultBelongsToJob(t *testing.T) {
	t.Run("match result belongs to job", func(t *testing.T) {
		repo, mock, _ := setupJobRepositoryTest(t)
		ctx := context.Background()
		jobID := 1
		matchID := 100

		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(matchID, jobID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		belongs, err := repo.MatchResultBelongsToJob(ctx, matchID, jobID)

		assert.NoError(t, err)
		assert.True(t, belongs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("match result does not belong to job", func(t *testing.T) {
		repo, mock, _ := setupJobRepositoryTest(t)
		ctx := context.Background()
		jobID := 1
		matchID := 100

		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(matchID, jobID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		belongs, err := repo.MatchResultBelongsToJob(ctx, matchID, jobID)

		assert.NoError(t, err)
		assert.False(t, belongs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid match ID", func(t *testing.T) {
		repo, _, _ := setupJobRepositoryTest(t)
		ctx := context.Background()

		belongs, err := repo.MatchResultBelongsToJob(ctx, 0, 1)

		assert.Equal(t, models.ErrInvalidJobID, err)
		assert.False(t, belongs)
	})

	t.Run("invalid job ID", func(t *testing.T) {
		repo, _, _ := setupJobRepositoryTest(t)
		ctx := context.Background()

		belongs, err := repo.MatchResultBelongsToJob(ctx, 1, 0)

		assert.Equal(t, models.ErrInvalidJobID, err)
		assert.False(t, belongs)
	})

	t.Run("non-existent match result", func(t *testing.T) {
		repo, mock, _ := setupJobRepositoryTest(t)
		ctx := context.Background()
		jobID := 999
		matchID := 999

		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(matchID, jobID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		belongs, err := repo.MatchResultBelongsToJob(ctx, matchID, jobID)

		assert.NoError(t, err)
		assert.False(t, belongs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
