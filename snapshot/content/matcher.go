package content

import (
	"sync"

	"github.com/tanemahuta/goselfie/utils"
)

// StringEncoder converts snapshot bytes to their serialized string form.
type StringEncoder interface {
	// Encode serializes contents for storage in the snapshot file.
	Encode(contents []byte) (string, error)
}

// Matcher identifies snapshot bytes and encodes them for storage.
type Matcher interface {
	utils.Prioritized
	StringEncoder
	// ContentType returns the type identified by this matcher.
	ContentType() Type
	// Matches reports whether contents have this matcher's content type.
	Matches(contents []byte) bool
}

// Registry detects content types and resolves their string encoders.
// Its zero value is ready to use; unmatched encoder lookups use BinaryContent.
type Registry struct {
	mu       sync.RWMutex
	matchers []Matcher
	encoders map[Type]StringEncoder
}

// NewRegistry creates an empty content registry with hexadecimal encoding as fallback.
func NewRegistry() *Registry {
	return &Registry{encoders: make(map[Type]StringEncoder)}
}

// Register adds a matcher and uses it as the encoder for its content type.
// Register ignores nil receivers and nil matchers. Registering another matcher
// for the same type replaces that type's encoder while keeping both matchers
// available for detection.
func (registry *Registry) Register(matcher Matcher) {
	if registry == nil || matcher == nil {
		return
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if registry.encoders == nil {
		registry.encoders = make(map[Type]StringEncoder)
	}
	registry.matchers = append(registry.matchers, matcher)
	registry.encoders[matcher.ContentType()] = matcher
}

// Detect returns the content type of the first matching matcher, or Unknown.
// Matchers are checked in ascending priority order.
func (registry *Registry) Detect(contents []byte) Type {
	if registry == nil {
		return Unknown
	}
	registry.mu.RLock()
	matchers := append([]Matcher(nil), registry.matchers...)
	registry.mu.RUnlock()
	utils.SortByPriority(matchers)
	for _, matcher := range matchers {
		if matcher.Matches(contents) {
			return matcher.ContentType()
		}
	}
	return Unknown
}

// Encoder returns the content-specific encoder, or BinaryContent when no
// encoder is registered for contentType.
func (registry *Registry) Encoder(contentType Type) StringEncoder {
	if registry == nil {
		return BinaryContent{}
	}
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	if encoder := registry.encoders[contentType]; encoder != nil {
		return encoder
	}
	return BinaryContent{}
}

// Matchers is the process-wide registry of snapshot content matchers.
// The YAML, text, and binary matchers are registered automatically.
var Matchers = NewRegistry() //nolint:gochecknoglobals // Built-in matchers register on the process-wide registry.
