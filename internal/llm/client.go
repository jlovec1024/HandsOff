package llm

import (
	"fmt"

	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/pkg/crypto"
)

// NewClient creates a new LLM client based on provider
func NewClient(provider *model.LLMProvider, encryptionKey string) (Client, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider cannot be nil")
	}

	// Decrypt API key
	apiKey, err := crypto.DecryptString(provider.APIKey, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt API key: %w", err)
	}

	// Create config
	config := Config{
		BaseURL:     provider.BaseURL,
		APIKey:      apiKey,
		ModelName:   provider.Model,
		MaxTokens:   4096,
		Temperature: 0.7,
		Timeout:     60,
	}

	// All providers are OpenAI-compatible, use unified client
	return createClientFromRegistry("openai-compatible", config)
}
