package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Validator handles webhook signature validation
type Validator struct{}

// NewValidator creates a new webhook validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateGitLabSignature validates GitLab webhook signature
// GitLab sends X-Gitlab-Token header with the secret token
func (v *Validator) ValidateGitLabSignature(payload []byte, receivedToken, expectedSecret string) error {
	// GitLab uses simple token comparison (not HMAC)
	// The webhook secret is sent as-is in X-Gitlab-Token header
	if receivedToken == "" {
		return fmt.Errorf("missing X-Gitlab-Token header")
	}

	if expectedSecret == "" {
		// If no secret is configured, skip validation (for development)
		return nil
	}

	if receivedToken != expectedSecret {
		return fmt.Errorf("invalid webhook token")
	}

	return nil
}

// ValidateGitHubSignature validates GitHub webhook signature (for future use)
// GitHub uses HMAC-SHA256 with X-Hub-Signature-256 header
func (v *Validator) ValidateGitHubSignature(payload []byte, receivedSignature, secret string) error {
	if receivedSignature == "" {
		return fmt.Errorf("missing X-Hub-Signature-256 header")
	}

	if secret == "" {
		return fmt.Errorf("webhook secret not configured")
	}

	// Remove "sha256=" prefix
	if len(receivedSignature) < 7 {
		return fmt.Errorf("invalid signature format")
	}
	receivedHash := receivedSignature[7:]

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(receivedHash), []byte(expectedHash)) {
		return fmt.Errorf("invalid webhook signature")
	}

	return nil
}
