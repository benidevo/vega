package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// HandlerTestCase defines a standard test case structure for HTTP handlers
type HandlerTestCase struct {
	Name           string
	Method         string
	Path           string
	Body           interface{}
	FormData       map[string]string
	Headers        map[string]string
	Cookies        []*http.Cookie
	MockSetup      func()
	ExpectedStatus int
	ExpectedHeader map[string]string
	ExpectedCookie []CookieAssertion
	ExpectedToast  *ToastAssertion
	ValidateBody   func(*testing.T, *httptest.ResponseRecorder)
}

// ToastAssertion represents expected HTMX toast notification
type ToastAssertion struct {
	Message string
	Type    string
}

// CookieAssertion represents expected cookie values
type CookieAssertion struct {
	Name     string
	Value    string
	HttpOnly bool
	Secure   bool
	SameSite http.SameSite
}

// SetupTestRouter creates a gin router in test mode
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

// CreateTestRequest creates an HTTP request for testing
func CreateTestRequest(method, path string, body interface{}) (*http.Request, error) {
	var reqBody io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			reqBody = strings.NewReader(v)
		case []byte:
			reqBody = bytes.NewReader(v)
		default:
			jsonBytes, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			reqBody = bytes.NewReader(jsonBytes)
		}
	}

	req, err := http.NewRequest(method, path, reqBody)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// CreateFormRequest creates a form-encoded HTTP request
func CreateFormRequest(method, path string, formData map[string]string) (*http.Request, error) {
	form := url.Values{}
	for key, value := range formData {
		form.Add(key, value)
	}

	req, err := http.NewRequest(method, path, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

// RunHandlerTest executes a handler test case
func RunHandlerTest(t *testing.T, router *gin.Engine, tc HandlerTestCase) {
	t.Helper()

	// Setup mocks if provided
	if tc.MockSetup != nil {
		tc.MockSetup()
	}

	// Create request
	var req *http.Request
	var err error

	if tc.FormData != nil {
		req, err = CreateFormRequest(tc.Method, tc.Path, tc.FormData)
	} else {
		req, err = CreateTestRequest(tc.Method, tc.Path, tc.Body)
	}
	assert.NoError(t, err)

	// Set headers
	for key, value := range tc.Headers {
		req.Header.Set(key, value)
	}

	// Set cookies
	for _, cookie := range tc.Cookies {
		req.AddCookie(cookie)
	}

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert status code
	assert.Equal(t, tc.ExpectedStatus, w.Code, "unexpected status code")

	// Assert headers
	for key, expectedValue := range tc.ExpectedHeader {
		actualValue := w.Header().Get(key)
		assert.Equal(t, expectedValue, actualValue, "unexpected header value for %s", key)
	}

	// Assert cookies
	if tc.ExpectedCookie != nil {
		cookies := w.Result().Cookies()
		for _, expected := range tc.ExpectedCookie {
			found := false
			for _, cookie := range cookies {
				if cookie.Name == expected.Name {
					found = true
					if expected.Value != "" {
						assert.Equal(t, expected.Value, cookie.Value, "unexpected cookie value for %s", expected.Name)
					}
					assert.Equal(t, expected.HttpOnly, cookie.HttpOnly, "unexpected HttpOnly for cookie %s", expected.Name)
					assert.Equal(t, expected.Secure, cookie.Secure, "unexpected Secure for cookie %s", expected.Name)
					if expected.SameSite != 0 {
						assert.Equal(t, expected.SameSite, cookie.SameSite, "unexpected SameSite for cookie %s", expected.Name)
					}
					break
				}
			}
			assert.True(t, found, "expected cookie %s not found", expected.Name)
		}
	}

	// Assert HTMX toast if expected
	if tc.ExpectedToast != nil {
		toastMessage := w.Header().Get("X-Toast-Message")
		toastType := w.Header().Get("X-Toast-Type")
		assert.Equal(t, tc.ExpectedToast.Message, toastMessage, "unexpected toast message")
		assert.Equal(t, tc.ExpectedToast.Type, toastType, "unexpected toast type")
	}

	// Custom body validation
	if tc.ValidateBody != nil {
		tc.ValidateBody(t, w)
	}
}

// AssertJSONResponse asserts that the response contains expected JSON
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expected interface{}) {
	t.Helper()

	var actual interface{}
	err := json.Unmarshal(w.Body.Bytes(), &actual)
	assert.NoError(t, err, "failed to unmarshal response body")

	expectedJSON, err := json.Marshal(expected)
	assert.NoError(t, err)

	actualJSON, err := json.Marshal(actual)
	assert.NoError(t, err)

	assert.JSONEq(t, string(expectedJSON), string(actualJSON), "unexpected JSON response")
}

// AssertHTMLContains asserts that the HTML response contains expected content
func AssertHTMLContains(t *testing.T, w *httptest.ResponseRecorder, expected ...string) {
	t.Helper()

	body := w.Body.String()
	for _, exp := range expected {
		assert.Contains(t, body, exp, "expected HTML to contain: %s", exp)
	}
}

// AssertHTMLNotContains asserts that the HTML response does not contain certain content
func AssertHTMLNotContains(t *testing.T, w *httptest.ResponseRecorder, unexpected ...string) {
	t.Helper()

	body := w.Body.String()
	for _, unexp := range unexpected {
		assert.NotContains(t, body, unexp, "expected HTML to not contain: %s", unexp)
	}
}

// AssertTemplateData asserts specific values in gin.H template data
func AssertTemplateData(t *testing.T, data gin.H, assertions map[string]interface{}) {
	t.Helper()

	for key, expected := range assertions {
		actual, exists := data[key]
		assert.True(t, exists, "expected template data to contain key: %s", key)
		assert.Equal(t, expected, actual, "unexpected value for template data key: %s", key)
	}
}

// CreateAuthenticatedRequest creates a request with authentication cookies
func CreateAuthenticatedRequest(method, path string, body interface{}, token string) (*http.Request, error) {
	req, err := CreateTestRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	cookie := &http.Cookie{
		Name:  "token",
		Value: token,
	}
	req.AddCookie(cookie)

	return req, nil
}

// MockHTMLRenderer provides a test renderer that captures template data
type MockHTMLRenderer struct {
	TemplateName   string
	TemplateData   gin.H
	ResponseStatus int
}

func (m *MockHTMLRenderer) HTML(c *gin.Context, code int, name string, obj interface{}) {
	m.ResponseStatus = code
	m.TemplateName = name
	if data, ok := obj.(gin.H); ok {
		m.TemplateData = data
	}
	c.Status(code)
}

// ExtractCSRFToken extracts CSRF token from HTML response
func ExtractCSRFToken(t *testing.T, body string) string {
	t.Helper()

	// Look for CSRF token in common patterns
	patterns := []string{
		`<input type="hidden" name="csrf_token" value="([^"]+)"`,
		`<meta name="csrf-token" content="([^"]+)"`,
		`data-csrf-token="([^"]+)"`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(body)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// MockServices provides a container for all mock services used in handler tests
type MockServices struct {
	AuthService    interface{}
	JobService     interface{}
	UserService    interface{}
	ProfileService interface{}
	QuotaService   interface{}
}

// SetupMockContext creates a gin context with common test setup
func SetupMockContext(w *httptest.ResponseRecorder) (*gin.Context, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	c, r := gin.CreateTestContext(w)
	return c, r
}
