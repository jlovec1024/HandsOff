package dto

// UpdatePlatformConfigRequest represents the request body for updating platform config
type UpdatePlatformConfigRequest struct {
	PlatformType string `json:"platform_type" binding:"required"`
	BaseURL      string `json:"base_url" binding:"required"`
	AccessToken  string `json:"access_token" binding:"required"`
	IsActive     bool   `json:"is_active"`
}
