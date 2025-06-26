package job

import (
	"context"
	"errors"
	"testing"

	"github.com/benidevo/vega/internal/job/models"
	"github.com/stretchr/testify/assert"
)

func TestJobService_DeleteMatchResult(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockJobRepository)
		cfg := setupTestConfig()
		service := NewJobService(mockRepo, nil, nil, cfg)
		ctx := context.Background()
		jobID := 1
		matchID := 100

		mockRepo.On("MatchResultBelongsToJob", ctx, matchID, jobID).Return(true, nil)
		mockRepo.On("DeleteMatchResult", ctx, matchID).Return(nil)

		err := service.DeleteMatchResult(ctx, jobID, matchID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid job ID", func(t *testing.T) {
		mockRepo := new(MockJobRepository)
		cfg := setupTestConfig()
		service := NewJobService(mockRepo, nil, nil, cfg)
		ctx := context.Background()

		err := service.DeleteMatchResult(ctx, 0, 100)

		assert.Equal(t, models.ErrInvalidJobID, err)
		mockRepo.AssertNotCalled(t, "MatchResultBelongsToJob")
		mockRepo.AssertNotCalled(t, "DeleteMatchResult")
	})

	t.Run("invalid match ID", func(t *testing.T) {
		mockRepo := new(MockJobRepository)
		cfg := setupTestConfig()
		service := NewJobService(mockRepo, nil, nil, cfg)
		ctx := context.Background()

		err := service.DeleteMatchResult(ctx, 1, 0)

		assert.Equal(t, models.ErrInvalidJobID, err)
		mockRepo.AssertNotCalled(t, "MatchResultBelongsToJob")
		mockRepo.AssertNotCalled(t, "DeleteMatchResult")
	})

	t.Run("match result does not belong to job", func(t *testing.T) {
		mockRepo := new(MockJobRepository)
		cfg := setupTestConfig()
		service := NewJobService(mockRepo, nil, nil, cfg)
		ctx := context.Background()
		jobID := 1
		matchID := 100

		mockRepo.On("MatchResultBelongsToJob", ctx, matchID, jobID).Return(false, nil)

		err := service.DeleteMatchResult(ctx, jobID, matchID)

		assert.Equal(t, models.ErrJobNotFound, err)
		mockRepo.AssertNotCalled(t, "DeleteMatchResult")
		mockRepo.AssertExpectations(t)
	})

	t.Run("error checking match ownership", func(t *testing.T) {
		mockRepo := new(MockJobRepository)
		cfg := setupTestConfig()
		service := NewJobService(mockRepo, nil, nil, cfg)
		ctx := context.Background()
		jobID := 1
		matchID := 100
		expectedErr := errors.New("database error")

		mockRepo.On("MatchResultBelongsToJob", ctx, matchID, jobID).Return(false, expectedErr)

		err := service.DeleteMatchResult(ctx, jobID, matchID)

		assert.Equal(t, expectedErr, err)
		mockRepo.AssertNotCalled(t, "DeleteMatchResult")
		mockRepo.AssertExpectations(t)
	})

	t.Run("error deleting match result", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockJobRepository)
		cfg := setupTestConfig()
		service := NewJobService(mockRepo, nil, nil, cfg)
		ctx := context.Background()
		jobID := 1
		matchID := 100
		expectedErr := errors.New("delete error")

		mockRepo.On("MatchResultBelongsToJob", ctx, matchID, jobID).Return(true, nil)
		mockRepo.On("DeleteMatchResult", ctx, matchID).Return(expectedErr)

		err := service.DeleteMatchResult(ctx, jobID, matchID)

		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}
