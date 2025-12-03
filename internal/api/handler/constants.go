package handler

// Pagination constants
const (
	// DefaultPage is the default page number for pagination
	DefaultPage = 1

	// DefaultPageSize is the default number of items per page
	DefaultPageSize = 20

	// MaxPageSize is the maximum allowed page size to prevent excessive database queries
	MaxPageSize = 100

	// MinPageSize is the minimum page size
	MinPageSize = 1
)

// Limit constants for query parameters
const (
	// DefaultRecentReviewsLimit is the default limit for recent reviews
	DefaultRecentReviewsLimit = 10

	// MaxRecentReviewsLimit is the maximum limit for recent reviews
	MaxRecentReviewsLimit = 100

	// MinRecentReviewsLimit is the minimum limit
	MinRecentReviewsLimit = 1
)

// Time range constants for dashboard queries
const (
	// DefaultTrendDays is the default number of days for trend data
	DefaultTrendDays = 7

	// MaxTrendDays is the maximum number of days for trend queries
	MaxTrendDays = 90

	// MinTrendDays is the minimum number of days
	MinTrendDays = 1

	// DefaultTrendDaysOnError fallback value when days parameter is out of range
	DefaultTrendDaysOnError = 30
)

// Error messages - Centralized for consistency
const (
	// ErrMsgInvalidRequest is returned when request body cannot be parsed
	ErrMsgInvalidRequest = "Invalid request"

	// ErrMsgInternalServer is a generic internal server error message
	ErrMsgInternalServer = "Internal server error"

	// ErrMsgProjectIDMissing is returned when project_id is missing from context
	ErrMsgProjectIDMissing = "Project ID missing from context - middleware failure"

	// ErrMsgUnauthorized is returned for authentication failures
	ErrMsgUnauthorized = "User not authenticated"

	// ErrMsgForbidden is returned for authorization failures
	ErrMsgForbidden = "Access forbidden"

	// ErrMsgNotFound is a generic not found error
	ErrMsgNotFound = "Resource not found"
)
