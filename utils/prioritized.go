package utils

import (
	"cmp"
	"slices"
)

// Prioritized entities.
type Prioritized interface {
	// Priority returns the priority of the item, lower values indicate higher priority.
	Priority() int
}

// SortByPriority sorts a slice of prioritized items in ascending order of their priority and returns them.
func SortByPriority[T Prioritized](items []T) []T {
	slices.SortFunc(items, func(lhs, rhs T) int {
		return cmp.Compare(lhs.Priority(), rhs.Priority())
	})
	return items
}
