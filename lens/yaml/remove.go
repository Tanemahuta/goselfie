package yaml

// Remove deletes the value at its path.
func Remove(path string) Transformer {
	return pathTransformer(path, func(any) (any, error) { return RemoveValue, nil })
}
