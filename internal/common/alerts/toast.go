package alerts

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
)

// ToastEvent represents the data structure for a toast notification event
type ToastEvent struct {
	ShowToast ToastData `json:"showToast"`
}

// ToastData contains the toast notification details
type ToastData struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// TriggerToast sets HTMX headers to trigger a toast notification on the client
func TriggerToast(c *gin.Context, message string, toastType AlertType) {
	event := ToastEvent{
		ShowToast: ToastData{
			Message: message,
			Type:    string(toastType),
		},
	}
	
	if data, err := json.Marshal(event); err == nil {
		c.Header("HX-Trigger", string(data))
	}
}

// TriggerToastAfterSwap sets HTMX headers to trigger a toast notification after content swap
func TriggerToastAfterSwap(c *gin.Context, message string, toastType AlertType) {
	event := ToastEvent{
		ShowToast: ToastData{
			Message: message,
			Type:    string(toastType),
		},
	}
	
	if data, err := json.Marshal(event); err == nil {
		c.Header("HX-Trigger-After-Swap", string(data))
	}
}

// TriggerToastAfterSettle sets HTMX headers to trigger a toast notification after content settles
func TriggerToastAfterSettle(c *gin.Context, message string, toastType AlertType) {
	event := ToastEvent{
		ShowToast: ToastData{
			Message: message,
			Type:    string(toastType),
		},
	}
	
	if data, err := json.Marshal(event); err == nil {
		c.Header("HX-Trigger-After-Settle", string(data))
	}
}