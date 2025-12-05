package llm

import (
	"fmt"
	"sync"

	"github.com/handsoff/handsoff/internal/model"
)

// ClientPool manages a pool of reusable LLM clients
// Thread-safe singleton pattern with lazy initialization
type ClientPool struct {
	clients sync.Map // key: string (providerID_modelID), value: Client
}

// Global client pool instance
var globalPool = &ClientPool{}

// GetOrCreateClient retrieves an existing client from pool or creates a new one
// This reduces overhead from repeated HTTP client creation and API key decryption
//
// Key format: "{providerID}_{modelID}" ensures configuration uniqueness
// Thread-safe: Multiple goroutines can safely call this concurrently
func GetOrCreateClient(provider *model.LLMProvider, encryptionKey string) (Client, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider cannot be nil")
	}

	// Generate cache key based on provider ID
	key := fmt.Sprintf("%d", provider.ID)

	// Try to get existing client from pool
	if cached, ok := globalPool.clients.Load(key); ok {
		return cached.(Client), nil
	}

	// Client not in pool, create new one
	client, err := NewClient(provider, encryptionKey)
	if err != nil {
		return nil, err
	}

	// Store in pool for future reuse
	globalPool.clients.Store(key, client)

	return client, nil
}

// InvalidateClient removes a client from the pool
// Should be called when provider/model configuration is updated or deleted
func InvalidateClient(providerID, modelID uint) {
	key := fmt.Sprintf("%d_%d", providerID, modelID)
	globalPool.clients.Delete(key)
}

// InvalidateProvider removes all clients associated with a provider
// Should be called when provider is updated (e.g., API key changed) or deleted
func InvalidateProvider(providerID uint) {
	// Iterate through all cached clients
	globalPool.clients.Range(func(key, value interface{}) bool {
		// Key format: "{providerID}_{modelID}"
		// Check if key starts with providerID
		keyStr := key.(string)
		expectedPrefix := fmt.Sprintf("%d_", providerID)
		
		// Delete if this client belongs to the provider
		if len(keyStr) >= len(expectedPrefix) && keyStr[:len(expectedPrefix)] == expectedPrefix {
			globalPool.clients.Delete(key)
		}
		
		return true // Continue iteration
	})
}

// ClearPool removes all clients from the pool
// Mainly for testing purposes
func ClearPool() {
	globalPool.clients.Range(func(key, value interface{}) bool {
		globalPool.clients.Delete(key)
		return true
	})
}
