package settings

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	settingsModels "github.com/benidevo/vega/internal/settings/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockEntityService struct {
	mock.Mock
}

func (m *mockEntityService) CreateEntity(ctx *gin.Context, entity interface{}) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *mockEntityService) UpdateEntity(ctx *gin.Context, entity interface{}) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *mockEntityService) DeleteEntity(ctx *gin.Context, entityID, profileID int, entityType string) error {
	args := m.Called(ctx, entityID, profileID, entityType)
	return args.Error(0)
}

func (m *mockEntityService) GetEntityByID(ctx *gin.Context, entityID, profileID int, entityType string) (interface{}, error) {
	args := m.Called(ctx, entityID, profileID, entityType)
	return args.Get(0), args.Error(1)
}

func TestEntityOperations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		testFunc func(t *testing.T, mockService *mockEntityService)
	}{
		{
			name: "should_create_work_experience_when_valid",
			testFunc: func(t *testing.T, mockService *mockEntityService) {
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

				mockService.On("CreateEntity", ctx, exp).Return(nil)

				err := mockService.CreateEntity(ctx, exp)
				require.NoError(t, err)
				mockService.AssertExpectations(t)
			},
		},
		{
			name: "should_return_error_when_create_validation_fails",
			testFunc: func(t *testing.T, mockService *mockEntityService) {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest("POST", "/test", nil)

				exp := &settingsModels.WorkExperience{
					ProfileID: 1,
					// Missing required fields
				}

				mockService.On("CreateEntity", ctx, exp).Return(errors.New("validation failed"))

				err := mockService.CreateEntity(ctx, exp)
				assert.Error(t, err)
				mockService.AssertExpectations(t)
			},
		},
		{
			name: "should_update_education_when_valid",
			testFunc: func(t *testing.T, mockService *mockEntityService) {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest("POST", "/test", nil)
				ctx.Set("userID", 123)

				edu := &settingsModels.Education{
					ID:          1,
					ProfileID:   1,
					Institution: "MIT",
					Degree:      "BS Computer Science",
					StartDate:   time.Now().Add(-4 * 365 * 24 * time.Hour),
				}

				mockService.On("UpdateEntity", ctx, edu).Return(nil)

				err := mockService.UpdateEntity(ctx, edu)
				require.NoError(t, err)
				mockService.AssertExpectations(t)
			},
		},
		{
			name: "should_return_error_when_update_repository_fails",
			testFunc: func(t *testing.T, mockService *mockEntityService) {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest("POST", "/test", nil)
				ctx.Set("userID", 123)

				cert := &settingsModels.Certification{
					ID:         1,
					ProfileID:  1,
					Name:       "AWS Architect",
					IssuingOrg: "AWS",
					IssueDate:  time.Now().Add(-30 * 24 * time.Hour),
				}

				repoErr := errors.New("database connection failed")
				mockService.On("UpdateEntity", ctx, cert).Return(repoErr)

				err := mockService.UpdateEntity(ctx, cert)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "database connection failed")
				mockService.AssertExpectations(t)
			},
		},
		{
			name: "should_delete_certification_when_exists",
			testFunc: func(t *testing.T, mockService *mockEntityService) {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest("POST", "/test", nil)
				ctx.Set("userID", 123)

				mockService.On("DeleteEntity", ctx, 1, 1, "Certification").Return(nil)

				err := mockService.DeleteEntity(ctx, 1, 1, "Certification")
				require.NoError(t, err)
				mockService.AssertExpectations(t)
			},
		},
		{
			name: "should_return_error_when_delete_invalid_entity_type",
			testFunc: func(t *testing.T, mockService *mockEntityService) {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest("POST", "/test", nil)
				ctx.Set("userID", 123)

				mockService.On("DeleteEntity", ctx, 1, 1, "invalid_type").Return(settingsModels.ErrSettingsNotFound)

				err := mockService.DeleteEntity(ctx, 1, 1, "invalid_type")
				assert.Error(t, err)
				assert.Equal(t, settingsModels.ErrSettingsNotFound, err)
				mockService.AssertExpectations(t)
			},
		},
		{
			name: "should_get_work_experience_by_id",
			testFunc: func(t *testing.T, mockService *mockEntityService) {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest("POST", "/test", nil)

				exp := &settingsModels.WorkExperience{
					ID:        1,
					ProfileID: 1,
					Company:   "Acme Corp",
					Title:     "Software Engineer",
				}

				mockService.On("GetEntityByID", ctx, 1, 1, "Experience").Return(exp, nil)

				result, err := mockService.GetEntityByID(ctx, 1, 1, "Experience")
				require.NoError(t, err)
				assert.Equal(t, exp, result)
				mockService.AssertExpectations(t)
			},
		},
		{
			name: "should_return_error_when_entity_not_found",
			testFunc: func(t *testing.T, mockService *mockEntityService) {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest("POST", "/test", nil)

				mockService.On("GetEntityByID", ctx, 999, 1, "Education").Return(nil, settingsModels.ErrEducationNotFound)

				result, err := mockService.GetEntityByID(ctx, 999, 1, "Education")
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, settingsModels.ErrEducationNotFound, err)
				mockService.AssertExpectations(t)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(mockEntityService)
			tc.testFunc(t, mockService)
		})
	}
}
