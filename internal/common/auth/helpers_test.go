package auth

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func(*gin.Context)
		expectedID    int
		expectedFound bool
	}{
		{
			name: "should_return_user_id_when_valid_integer_in_context",
			setupContext: func(c *gin.Context) {
				c.Set("userID", 123)
			},
			expectedID:    123,
			expectedFound: true,
		},
		{
			name:          "should_return_false_when_user_id_not_set",
			setupContext:  func(c *gin.Context) {},
			expectedID:    0,
			expectedFound: false,
		},
		{
			name: "should_return_false_when_user_id_is_not_int",
			setupContext: func(c *gin.Context) {
				c.Set("userID", "123")
			},
			expectedID:    0,
			expectedFound: false,
		},
		{
			name: "should_return_false_when_user_id_is_nil",
			setupContext: func(c *gin.Context) {
				c.Set("userID", nil)
			},
			expectedID:    0,
			expectedFound: false,
		},
		{
			name: "should_return_false_when_user_id_is_float",
			setupContext: func(c *gin.Context) {
				c.Set("userID", 123.45)
			},
			expectedID:    0,
			expectedFound: false,
		},
		{
			name: "should_handle_zero_user_id_when_zero_value_in_context",
			setupContext: func(c *gin.Context) {
				c.Set("userID", 0)
			},
			expectedID:    0,
			expectedFound: true,
		},
		{
			name: "should_handle_negative_user_id_when_negative_value_in_context",
			setupContext: func(c *gin.Context) {
				c.Set("userID", -1)
			},
			expectedID:    -1,
			expectedFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setupContext(c)

			id, found := GetUserID(c)

			assert.Equal(t, tt.expectedID, id)
			assert.Equal(t, tt.expectedFound, found)
		})
	}
}

func TestMustGetUserID(t *testing.T) {
	tests := []struct {
		name         string
		setupContext func(*gin.Context)
		expectedID   int
		shouldPanic  bool
	}{
		{
			name: "should_return_user_id_when_valid_integer_in_context",
			setupContext: func(c *gin.Context) {
				c.Set("userID", 456)
			},
			expectedID:  456,
			shouldPanic: false,
		},
		{
			name:         "should_panic_when_user_id_not_set",
			setupContext: func(c *gin.Context) {},
			shouldPanic:  true,
		},
		{
			name: "should_panic_when_user_id_is_not_int",
			setupContext: func(c *gin.Context) {
				c.Set("userID", "456")
			},
			shouldPanic: true,
		},
		{
			name: "should_panic_when_user_id_is_nil",
			setupContext: func(c *gin.Context) {
				c.Set("userID", nil)
			},
			shouldPanic: true,
		},
		{
			name: "should_handle_zero_user_id_when_zero_value_in_context",
			setupContext: func(c *gin.Context) {
				c.Set("userID", 0)
			},
			expectedID:  0,
			shouldPanic: false,
		},
		{
			name: "should_handle_negative_user_id_when_negative_value_in_context",
			setupContext: func(c *gin.Context) {
				c.Set("userID", -99)
			},
			expectedID:  -99,
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setupContext(c)

			if tt.shouldPanic {
				assert.PanicsWithValue(t, "userID not found in context or invalid type", func() {
					MustGetUserID(c)
				})
			} else {
				assert.NotPanics(t, func() {
					id := MustGetUserID(c)
					assert.Equal(t, tt.expectedID, id)
				})
			}
		})
	}
}

func TestGetUserIDWithMultipleContextValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("userID", 789)
	c.Set("username", "testuser")
	c.Set("role", "admin")

	id, found := GetUserID(c)
	assert.True(t, found)
	assert.Equal(t, 789, id)
}

func TestGetUserIDThreadSafety(t *testing.T) {
	gin.SetMode(gin.TestMode)

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("userID", idx)

			id, found := GetUserID(c)
			assert.True(t, found)
			assert.Equal(t, idx, id)

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
