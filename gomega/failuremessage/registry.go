package failuremessage

import (
	"sync"

	"github.com/tanemahuta/goselfie/snapshot/content"
)

// Registry maps content types to failure message providers.
type Registry struct {
	mu              sync.RWMutex
	providers       map[content.Type]Provider
	defaultProvider Provider
}

// NewRegistry creates a provider registry with the binary provider as fallback.
func NewRegistry() *Registry {
	return NewRegistryWithDefault(binaryProvider{})
}

// NewRegistryWithDefault creates a provider registry with a custom fallback.
func NewRegistryWithDefault(defaultProvider Provider) *Registry {
	if defaultProvider == nil {
		defaultProvider = binaryProvider{}
	}
	return &Registry{
		providers:       make(map[content.Type]Provider),
		defaultProvider: defaultProvider,
	}
}

// Register installs provider for contentType, replacing any previous provider.
func (registry *Registry) Register(contentType content.Type, provider Provider) {
	if registry == nil || provider == nil {
		return
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.providers[contentType] = provider
}

// Provider returns the provider registered for contentType, or the default provider.
func (registry *Registry) Provider(contentType content.Type) Provider {
	if registry == nil {
		return nil
	}
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	if provider := registry.providers[contentType]; provider != nil {
		return provider
	}
	return registry.defaultProvider
}

// DefaultProvider returns the fallback provider used for unregistered content types.
func (registry *Registry) DefaultProvider() Provider {
	if registry == nil {
		return nil
	}
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	return registry.defaultProvider
}

// Providers is the process-wide failure message provider registry.
var Providers = NewRegistry() //nolint:gochecknoglobals // Applications can register providers on the default registry.

// For returns the provider registered for contentType, or the registry's default.
func For(contentType content.Type) Provider {
	return Providers.Provider(contentType)
}
