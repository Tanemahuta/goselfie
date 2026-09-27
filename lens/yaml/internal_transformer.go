package yaml

import "github.com/tanemahuta/goselfie/utils/path"

// internalTransformer checks its path before changing a value.
type internalTransformer struct {
	matcher     path.Matcher
	transformer ValueTransformer
	priority    int
}

// NewTransformer creates a path-aware transform. A nil matcher matches all paths.
func NewTransformer(matcher path.Matcher, transformer ValueTransformer) Transformer {
	return internalTransformer{matcher: matcher, transformer: transformer}
}

func (transform internalTransformer) Apply(path []string, value any) (any, error) {
	if transform.matcher != nil {
		if !transform.matcher(path) {
			return value, nil
		}
	}
	if transform.transformer == nil {
		return value, nil
	}
	return transform.transformer(value)
}
func (transform internalTransformer) Priority() int { return transform.priority }
func (transform internalTransformer) WithPriority(priority int) Transformer {
	transform.priority = priority
	return transform
}
