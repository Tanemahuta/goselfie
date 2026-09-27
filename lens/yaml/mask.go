package yaml

// Mask replaces the value at its path with a stable marker.
func Mask(path string) Transformer {
	return pathTransformer(path, func(any) (any, error) { return "<masked>", nil })
}
