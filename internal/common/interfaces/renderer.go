package interfaces

import "github.com/gin-gonic/gin"

// HTMLRenderer handles HTML template rendering
type HTMLRenderer interface {
	HTML(c *gin.Context, code int, name string, obj interface{})
}

// ErrorHandler handles error responses
type ErrorHandler interface {
	HandleError(c *gin.Context, err error)
	HandleValidationError(c *gin.Context, err error)
}
