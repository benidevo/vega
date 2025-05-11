package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testAuthHandler struct {
	AuthHandler
	testService *testAuthService
}

func newTestAuthHandler() *testAuthHandler {
	ts := &testAuthService{}
	return &testAuthHandler{
		testService: ts,
	}
}

func (h *testAuthHandler) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	token, loginErr := h.testService.Login(c.Request.Context(), username, password)
	if loginErr != nil {
		errorHTML := `<div id="form-response" class="text-center text-sm mb-3 p-2 bg-red-800 bg-opacity-50 text-white rounded-md">
			` + loginErr.Error() + `
		</div>`

		c.Header("HX-Reswap", "innerHTML")
		c.Header("HX-Retarget", "#form-response")
		c.Data(http.StatusUnauthorized, "text/html", []byte(errorHTML))
		return
	}

	c.SetCookie("token", token, 3600, "/", "", false, true)
	c.Header("HX-Redirect", "/dashboard")
	c.Status(http.StatusOK)
}

func (h *testAuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil {
			c.Redirect(http.StatusFound, "/auth/login")
			c.Abort()
			return
		}

		claims, err := h.testService.VerifyToken(tokenString)
		if err != nil {
			c.Redirect(http.StatusFound, "/auth/login")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func TestLoginHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should_return_error_when_login_fails", func(t *testing.T) {
		r := gin.New()

		handler := newTestAuthHandler()
		handler.testService.loginError = ErrInvalidCredentials

		r.POST("/login", handler.Login)

		form := url.Values{}
		form.Add("username", "testuser")
		form.Add("password", "wrongpassword")

		req, _ := http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid credentials")
		assert.Equal(t, "innerHTML", w.Header().Get("HX-Reswap"))
		assert.Equal(t, "#form-response", w.Header().Get("HX-Retarget"))
	})

	t.Run("should_set_cookie_and_redirect_when_login_successful", func(t *testing.T) {
		r := gin.New()

		handler := newTestAuthHandler()
		handler.testService.loginResult = "test.jwt.token"

		r.POST("/login", handler.Login)

		form := url.Values{}
		form.Add("username", "testuser")
		form.Add("password", "validpassword")

		req, _ := http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "/dashboard", w.Header().Get("HX-Redirect"))

		cookies := w.Result().Cookies()
		var tokenCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "token" {
				tokenCookie = cookie
				break
			}
		}
		require.NotNil(t, tokenCookie, "Token cookie should be set")
		assert.Equal(t, "test.jwt.token", tokenCookie.Value)
		assert.True(t, tokenCookie.HttpOnly, "Cookie should be HTTP-only")
	})
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should_redirect_to_login_when_no_token_provided", func(t *testing.T) {
		r := gin.New()

		handler := newTestAuthHandler()

		r.GET("/protected", handler.AuthMiddleware(), func(c *gin.Context) {
			c.String(http.StatusOK, "Protected content")
		})

		req, _ := http.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "/auth/login", w.Header().Get("Location"))
	})

	t.Run("should_redirect_to_login_when_token_is_invalid", func(t *testing.T) {
		r := gin.New()

		handler := newTestAuthHandler()
		handler.testService.verifyError = ErrInvalidToken

		r.GET("/protected", handler.AuthMiddleware(), func(c *gin.Context) {
			c.String(http.StatusOK, "Protected content")
		})

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "invalid.token"})
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "/auth/login", w.Header().Get("Location"))
	})

	t.Run("should_set_user_context_and_allow_access_when_token_is_valid", func(t *testing.T) {
		r := gin.New()

		handler := newTestAuthHandler()
		handler.testService.verifyClaims = &Claims{
			UserID:   1,
			Username: "testuser",
			Role:     "Admin",
		}

		r.GET("/protected", handler.AuthMiddleware(), func(c *gin.Context) {
			userID, exists := c.Get("userID")
			assert.True(t, exists, "userID should be set in context")
			assert.Equal(t, 1, userID.(int))

			username, exists := c.Get("username")
			assert.True(t, exists, "username should be set in context")
			assert.Equal(t, "testuser", username.(string))

			role, exists := c.Get("role")
			assert.True(t, exists, "role should be set in context")
			assert.Equal(t, "Admin", role.(string))

			c.String(http.StatusOK, "Protected content")
		})

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "valid.token"})
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Protected content", w.Body.String())
	})
}

type testAuthService struct {
	loginResult  string
	loginError   error
	verifyClaims *Claims
	verifyError  error
}

func (t *testAuthService) Login(ctx context.Context, username, password string) (string, error) {
	if t.loginError != nil {
		return "", t.loginError
	}
	return t.loginResult, nil
}

func (t *testAuthService) VerifyToken(token string) (*Claims, error) {
	if t.verifyError != nil {
		return nil, t.verifyError
	}
	return t.verifyClaims, nil
}
