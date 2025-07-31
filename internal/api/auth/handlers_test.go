package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benidevo/vega/internal/common/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockOAuthService implements the oauthService interface for testing
type mockOAuthService struct {
	mock.Mock
}

func (m *mockOAuthService) Authenticate(ctx context.Context, code, redirectURI string) (string, string, error) {
	args := m.Called(ctx, code, redirectURI)
	return args.String(0), args.String(1), args.Error(2)
}

// mockAuthService implements the authService interface for testing
type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Login(ctx context.Context, username, password string) (string, string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *mockAuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.Error(1)
}

func setupTestAuthAPIHandler() (*AuthAPIHandler, *mockOAuthService, *mockAuthService, *gin.Engine) {
	mockOAuth := new(mockOAuthService)
	mockAuth := new(mockAuthService)

	handler := NewAuthAPIHandler(mockOAuth, mockAuth)
	router := testutil.SetupTestRouter()

	return handler, mockOAuth, mockAuth, router
}

func TestAuthAPIHandler_ExchangeTokenForJWT(t *testing.T) {
	handler, mockOAuth, _, router := setupTestAuthAPIHandler()

	// Setup routes
	router.POST("/api/auth/token", handler.ExchangeTokenForJWT)

	tests := []testutil.HandlerTestCase{
		{
			Name:   "should_exchange_oauth_code_when_valid",
			Method: "POST",
			Path:   "/api/auth/token",
			Body: map[string]string{
				"code":         "valid-oauth-code",
				"redirect_uri": "http://localhost:3000/callback",
			},
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			MockSetup: func() {
				mockOAuth.On("Authenticate", mock.Anything, "valid-oauth-code", "http://localhost:3000/callback").
					Return("access-token", "refresh-token", nil)
			},
			ExpectedStatus: http.StatusOK,
			ValidateBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "access-token", response["token"])
				assert.Equal(t, "refresh-token", response["refresh_token"])
			},
		},
		{
			Name:   "should_return_unauthorized_when_invalid_code",
			Method: "POST",
			Path:   "/api/auth/token",
			Body: map[string]string{
				"code":         "invalid-oauth-code",
				"redirect_uri": "http://localhost:3000/callback",
			},
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			MockSetup: func() {
				mockOAuth.On("Authenticate", mock.Anything, "invalid-oauth-code", "http://localhost:3000/callback").
					Return("", "", errors.New("invalid authorization code"))
			},
			ExpectedStatus: http.StatusUnauthorized,
			ValidateBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "failed to exchange code for JWT", response["error"])
			},
		},
		{
			Name:   "should_return_bad_request_when_missing_code",
			Method: "POST",
			Path:   "/api/auth/token",
			Body: map[string]string{
				"redirect_uri": "http://localhost:3000/callback",
			},
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			ExpectedStatus: http.StatusBadRequest,
			ValidateBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "invalid request body", response["error"])
			},
		},
		{
			Name:   "should_return_bad_request_when_missing_redirect_uri",
			Method: "POST",
			Path:   "/api/auth/token",
			Body: map[string]string{
				"code": "valid-oauth-code",
			},
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			ExpectedStatus: http.StatusBadRequest,
			ValidateBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "invalid request body", response["error"])
			},
		},
		{
			Name:           "should_return_bad_request_when_invalid_json",
			Method:         "POST",
			Path:           "/api/auth/token",
			Body:           "invalid-json",
			Headers:        map[string]string{"Content-Type": "application/json"},
			ExpectedStatus: http.StatusBadRequest,
			ValidateBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "invalid request body", response["error"])
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			mockOAuth.ExpectedCalls = nil
			mockOAuth.Calls = nil
			testutil.RunHandlerTest(t, router, tc)
			mockOAuth.AssertExpectations(t)
		})
	}
}
