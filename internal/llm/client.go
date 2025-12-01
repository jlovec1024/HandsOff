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

	// Select client implementation based on provider type
	switch provider.Type {
	case "openai":
		return NewOpenAIClient(config), nil
	case "deepseek":
		return NewDeepSeekClient(config), nil
	case "claude":
		return nil, fmt.Errorf("claude provider not yet implemented")
	case "gemini":
		return nil, fmt.Errorf("gemini provider not yet implemented")
	case "ollama":
		return nil, fmt.Errorf("ollama provider not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", provider.Type)
	}
}
