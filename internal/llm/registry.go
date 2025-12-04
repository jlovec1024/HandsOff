package llm

import (
	"fmt"
	"sync"
)

// ClientFactory is a function that creates a Client from a Config
type ClientFactory func(config Config) Client

// providerRegistry stores registered provider factories
var providerRegistry = struct {
	sync.RWMutex
	factories map[string]ClientFactory
}{
	factories: make(map[string]ClientFactory),
}

// RegisterProvider registers a provider factory function
// This allows adding new providers without modifying core code
//
// Example:
//   RegisterProvider("openai", func(c Config) Client {
//       return NewOpenAICompatibleClient("OpenAI", c)
//   })
func RegisterProvider(providerType string, factory ClientFactory) {
	providerRegistry.Lock()
	defer providerRegistry.Unlock()
	providerRegistry.factories[providerType] = factory
}

// GetProviderFactory retrieves a registered provider factory
// Returns nil if provider type is not registered
func GetProviderFactory(providerType string) (ClientFactory, bool) {
	providerRegistry.RLock()
	defer providerRegistry.RUnlock()
	factory, exists := providerRegistry.factories[providerType]
	return factory, exists
}

// ListRegisteredProviders returns all registered provider types
func ListRegisteredProviders() []string {
	providerRegistry.RLock()
	defer providerRegistry.RUnlock()
	
	providers := make([]string, 0, len(providerRegistry.factories))
	for providerType := range providerRegistry.factories {
		providers = append(providers, providerType)
	}
	return providers
}

// init registers default providers
// This runs automatically when the package is imported
func init() {
	// Register OpenAI-compatible providers
	RegisterProvider("openai", func(c Config) Client {
		return NewOpenAICompatibleClient("OpenAI", c)
	})
	
	RegisterProvider("deepseek", func(c Config) Client {
		return NewOpenAICompatibleClient("DeepSeek", c)
	})
	
	// Future providers can be registered here or externally:
	// RegisterProvider("claude", func(c Config) Client {
	//     return NewClaudeClient(c)
	// })
}

// createClientFromRegistry creates a client using the registry
// Returns an error if provider type is not registered
func createClientFromRegistry(providerType string, config Config) (Client, error) {
	factory, exists := GetProviderFactory(providerType)
	if !exists {
		return nil, fmt.Errorf("unsupported provider type: %s (registered: %v)", 
			providerType, ListRegisteredProviders())
	}
	
	return factory(config), nil
}
