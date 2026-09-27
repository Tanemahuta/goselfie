package utils

import "cmp"

// Range represents a half-open range of ordered values.
type Range[T cmp.Ordered] struct {
	// Start of range, inclusive.
	Start T
	// End of range, exclusive.
	EndExcl T
}

// Contains reports whether value is within the range.
func (r Range[T]) Contains(value T) bool {
	return r.Start <= value && value < r.EndExcl
}
