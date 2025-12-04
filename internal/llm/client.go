package llm

import (
	"fmt"

	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/pkg/crypto"
)

// NewClient creates a new LLM client based on provider type
func NewClient(provider *model.LLMProvider, llmModel *model.LLMModel, encryptionKey string) (Client, error) {
	if provider == nil || llmModel == nil {
		return nil, fmt.Errorf("provider and model cannot be nil")
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
		ModelName:   llmModel.ModelName,
		MaxTokens:   llmModel.MaxTokens,
		Temperature: llmModel.Temperature,
		Timeout:     60, // 60 seconds default
	}

	// Use registry to create client (Open-Closed Principle)
	// New providers can be registered without modifying this code
	return createClientFromRegistry(provider.Type, config)
}
