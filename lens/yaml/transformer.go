package yaml

import "strings"

// Transformer changes a plain YAML value at path. Return RemoveValue to remove
// the value. Returning nil keeps a YAML null value.
type Transformer interface {
	// Apply transforms value when its path matches.
	Apply(path []string, value any) (any, error)
	// Priority returns the transform order; lower priorities run first.
	Priority() int
	// WithPriority returns an independent transformer with priority.
	WithPriority(priority int) Transformer
}

// ValueTransformer changes a matched YAML value.
type ValueTransformer func(any) (any, error)

// RemoveValue marks a value for removal. Nil remains a valid value.
var RemoveValue = struct{}{}

// IsRemove reports whether value is the removal marker.
func IsRemove(value any) bool {
	marker, ok := value.(struct{})
	return ok && marker == RemoveValue
}

func formatPath(path []string) string {
	if len(path) == 0 {
		return "<root>"
	}
	return strings.Join(path, ".")
}
