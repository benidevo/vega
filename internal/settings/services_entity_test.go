package settings

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	settingsModels "github.com/benidevo/vega/internal/settings/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityOperations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service, mockProfileRepo, _ := setupTestService()

	t.Run("CreateEntity: work experience", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/test", nil)

		exp := &settingsModels.WorkExperience{
			ProfileID:   1,
			Company:     "Acme Corp",
			Title:       "Software Engineer",
			StartDate:   time.Now().Add(-365 * 24 * time.Hour),
			Description: "Building software",
		}

		mockProfileRepo.On("AddWorkExperience", context.Background(), exp).Return(nil).Once()

		err := service.CreateEntity(ctx, exp)
		require.NoError(t, err)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("CreateEntity - validation failure", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/test", nil)
		exp := &settingsModels.WorkExperience{
			// Missing required fields
			ProfileID: 1,
		}

		err := service.CreateEntity(ctx, exp)
		assert.Error(t, err)
		mockProfileRepo.AssertNotCalled(t, "AddWorkExperience")
	})

	t.Run("UpdateEntity: education", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/test", nil)
		edu := &settingsModels.Education{
			ID:          1,
			ProfileID:   1,
			Institution: "MIT",
			Degree:      "BS Computer Science",
			StartDate:   time.Now().Add(-4 * 365 * 24 * time.Hour),
		}

		mockProfileRepo.On("UpdateEducation", context.Background(), edu).Return(edu, nil).Once()

		err := service.UpdateEntity(ctx, edu)
		require.NoError(t, err)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("UpdateEntity: repository error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/test", nil)
		cert := &settingsModels.Certification{
			ID:         1,
			ProfileID:  1,
			Name:       "AWS Architect",
			IssuingOrg: "AWS",
			IssueDate:  time.Now().Add(-30 * 24 * time.Hour),
		}

		repoErr := errors.New("database connection failed")
		mockProfileRepo.On("UpdateCertification", context.Background(), cert).Return(nil, repoErr).Once()

		err := service.UpdateEntity(ctx, cert)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection failed")
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("DeleteEntity: certification", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/test", nil)

		cert := &settingsModels.Certification{
			ID:         1,
			ProfileID:  1,
			Name:       "AWS Architect",
			IssuingOrg: "AWS",
		}

		// DeleteEntity first calls GetEntityByID to verify entity exists
		mockProfileRepo.On("GetEntityByID", context.Background(), 1, 1, "Certification").Return(cert, nil).Once()
		mockProfileRepo.On("DeleteCertification", context.Background(), 1).Return(nil).Once()

		err := service.DeleteEntity(ctx, 1, 1, "Certification")
		require.NoError(t, err)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("DeleteEntity: invalid entity type", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/test", nil)

		// GetEntityByID will be called first and should return an error for invalid type
		mockProfileRepo.On("GetEntityByID", context.Background(), 1, 1, "invalid_type").Return(nil, settingsModels.ErrSettingsNotFound).Once()

		err := service.DeleteEntity(ctx, 1, 1, "invalid_type")
		assert.Error(t, err)
		assert.Equal(t, settingsModels.ErrSettingsNotFound, err)
		mockProfileRepo.AssertNotCalled(t, "DeleteWorkExperience")
		mockProfileRepo.AssertNotCalled(t, "DeleteEducation")
		mockProfileRepo.AssertNotCalled(t, "DeleteCertification")
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("GetEntityByID: work experience", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/test", nil)
		exp := &settingsModels.WorkExperience{
			ID:        1,
			ProfileID: 1,
			Company:   "Acme Corp",
			Title:     "Software Engineer",
		}

		mockProfileRepo.On("GetEntityByID", context.Background(), 1, 1, "Experience").Return(exp, nil).Once()

		result, err := service.GetEntityByID(ctx, 1, 1, "Experience")
		require.NoError(t, err)
		assert.Equal(t, exp, result)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("GetEntityByID: entity not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/test", nil)

		mockProfileRepo.On("GetEntityByID", context.Background(), 999, 1, "Education").Return(nil, settingsModels.ErrEducationNotFound).Once()

		result, err := service.GetEntityByID(ctx, 999, 1, "Education")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, settingsModels.ErrEducationNotFound, err)
		mockProfileRepo.AssertExpectations(t)
	})
}
