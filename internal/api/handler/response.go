package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`            // Human-readable error message
	Code    string `json:"code,omitempty"`   // Optional: Machine-readable error code
	Details string `json:"details,omitempty"` // Optional: Additional details
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Message string      `json:"message,omitempty"` // Optional success message
	Data    interface{} `json:"data,omitempty"`    // Optional response data
}

// Response helper functions for consistent API responses

// RespondError sends a standardized error response
func RespondError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{
		Error: message,
	})
}

// RespondErrorWithCode sends an error response with error code
func RespondErrorWithCode(c *gin.Context, statusCode int, message, code string) {
	c.JSON(statusCode, ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// RespondErrorWithDetails sends an error response with additional details
func RespondErrorWithDetails(c *gin.Context, statusCode int, message, details string) {
	c.JSON(statusCode, ErrorResponse{
		Error:   message,
		Details: details,
	})
}

// RespondSuccess sends a standardized success response with data
func RespondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// RespondSuccessWithMessage sends a success response with message and optional data
func RespondSuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Message: message,
		Data:    data,
	})
}

// RespondCreated sends a 201 Created response
func RespondCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// RespondNoContent sends a 204 No Content response
func RespondNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Common error response helpers

// RespondBadRequest sends a 400 Bad Request error
func RespondBadRequest(c *gin.Context, message string) {
	RespondError(c, http.StatusBadRequest, message)
}

// RespondUnauthorized sends a 401 Unauthorized error
func RespondUnauthorized(c *gin.Context, message string) {
	RespondError(c, http.StatusUnauthorized, message)
}

// RespondForbidden sends a 403 Forbidden error
func RespondForbidden(c *gin.Context, message string) {
	RespondError(c, http.StatusForbidden, message)
}

// RespondNotFound sends a 404 Not Found error
func RespondNotFound(c *gin.Context, message string) {
	RespondError(c, http.StatusNotFound, message)
}

// RespondInternalError sends a 500 Internal Server Error
func RespondInternalError(c *gin.Context, message string) {
	RespondError(c, http.StatusInternalServerError, message)
}

// RespondValidationError sends a 400 Bad Request for validation errors
func RespondValidationError(c *gin.Context, details string) {
	RespondErrorWithDetails(c, http.StatusBadRequest, "Validation failed", details)
}
