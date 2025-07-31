package alerts

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTriggerToast(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		toastType    AlertType
		expectedJSON string
	}{
		{
			name:         "should_set_error_header_when_error_type",
			message:      "Error occurred",
			toastType:    TypeError,
			expectedJSON: `{"showToast":{"message":"Error occurred","type":"error"}}`,
		},
		{
			name:         "should_set_success_header_when_success_type",
			message:      "Operation successful",
			toastType:    TypeSuccess,
			expectedJSON: `{"showToast":{"message":"Operation successful","type":"success"}}`,
		},
		{
			name:         "should_set_warning_header_when_warning_type",
			message:      "Warning message",
			toastType:    TypeWarning,
			expectedJSON: `{"showToast":{"message":"Warning message","type":"warning"}}`,
		},
		{
			name:         "should_set_info_header_when_info_type",
			message:      "Information",
			toastType:    TypeInfo,
			expectedJSON: `{"showToast":{"message":"Information","type":"info"}}`,
		},
		{
			name:         "should_set_header_when_empty_message",
			message:      "",
			toastType:    TypeInfo,
			expectedJSON: `{"showToast":{"message":"","type":"info"}}`,
		},
		{
			name:         "should_escape_json_when_special_characters_in_message",
			message:      `Message with "quotes" and 'apostrophes'`,
			toastType:    TypeError,
			expectedJSON: `{"showToast":{"message":"Message with \"quotes\" and 'apostrophes'","type":"error"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			TriggerToast(c, tt.message, tt.toastType)

			assert.Equal(t, tt.expectedJSON, w.Header().Get("HX-Trigger"))
		})
	}
}

func TestTriggerToastAfterSwap(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		toastType    AlertType
		expectedJSON string
	}{
		{
			name:         "should_set_after_swap_header_when_success_type",
			message:      "Content swapped",
			toastType:    TypeSuccess,
			expectedJSON: `{"showToast":{"message":"Content swapped","type":"success"}}`,
		},
		{
			name:         "should_set_after_swap_header_when_error_type",
			message:      "Swap failed",
			toastType:    TypeError,
			expectedJSON: `{"showToast":{"message":"Swap failed","type":"error"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			TriggerToastAfterSwap(c, tt.message, tt.toastType)

			assert.Equal(t, tt.expectedJSON, w.Header().Get("HX-Trigger-After-Swap"))
		})
	}
}

func TestTriggerToastAfterSettle(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		toastType    AlertType
		expectedJSON string
	}{
		{
			name:         "should_set_after_settle_header_when_info_type",
			message:      "Content settled",
			toastType:    TypeInfo,
			expectedJSON: `{"showToast":{"message":"Content settled","type":"info"}}`,
		},
		{
			name:         "should_set_after_settle_header_when_warning_type",
			message:      "Settlement warning",
			toastType:    TypeWarning,
			expectedJSON: `{"showToast":{"message":"Settlement warning","type":"warning"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			TriggerToastAfterSettle(c, tt.message, tt.toastType)

			assert.Equal(t, tt.expectedJSON, w.Header().Get("HX-Trigger-After-Settle"))
		})
	}
}

func TestToastEventStructure(t *testing.T) {
	event := ToastEvent{
		ShowToast: ToastData{
			Message: "Test message",
			Type:    "success",
		},
	}

	data, err := json.Marshal(event)
	assert.NoError(t, err)

	expected := `{"showToast":{"message":"Test message","type":"success"}}`
	assert.JSONEq(t, expected, string(data))

	var decoded ToastEvent
	err = json.Unmarshal([]byte(expected), &decoded)
	assert.NoError(t, err)
	assert.Equal(t, event.ShowToast.Message, decoded.ShowToast.Message)
	assert.Equal(t, event.ShowToast.Type, decoded.ShowToast.Type)
}

func TestMultipleToastTriggers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	TriggerToast(c, "First message", TypeError)
	TriggerToastAfterSwap(c, "Second message", TypeSuccess)
	TriggerToastAfterSettle(c, "Third message", TypeInfo)

	assert.Equal(t, `{"showToast":{"message":"First message","type":"error"}}`, w.Header().Get("HX-Trigger"))
	assert.Equal(t, `{"showToast":{"message":"Second message","type":"success"}}`, w.Header().Get("HX-Trigger-After-Swap"))
	assert.Equal(t, `{"showToast":{"message":"Third message","type":"info"}}`, w.Header().Get("HX-Trigger-After-Settle"))
}

func TestToastWithLongMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	longMessage := "This is a very long message that contains a lot of text to test how the system handles lengthy toast notifications that might wrap or need special handling in the UI"

	TriggerToast(c, longMessage, TypeWarning)

	header := w.Header().Get("HX-Trigger")
	assert.Contains(t, header, longMessage)
	assert.Contains(t, header, "warning")
}
