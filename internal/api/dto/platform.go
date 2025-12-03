package dto

// UpdatePlatformConfigRequest represents the request body for updating platform config
type UpdatePlatformConfigRequest struct {
	PlatformType string  `json:"platform_type" binding:"required"`
	BaseURL      string  `json:"base_url" binding:"required"`
	AccessToken  *string `json:"access_token"` // Pointer: nil = keep existing, non-nil = update
	IsActive     bool    `json:"is_active"`
}

// TestConnectionRequest represents the request body for testing platform connection
// This allows users to test configuration before saving
type TestConnectionRequest struct {
	PlatformType string `json:"platform_type" binding:"required"`
	BaseURL      string `json:"base_url" binding:"required"`
	AccessToken  string `json:"access_token" binding:"required"`
}
