// Package yaml provides path-aware transformations for YAML values.
package yaml

import (
	"cmp"
	"fmt"
	"slices"
)

// Combine applies transforms in ascending priority order at one node.
func Combine(transforms ...Transformer) Transformer {
	flat := make([]Transformer, 0, len(transforms))
	for _, transform := range transforms {
		if transform == nil {
			continue
		}
		if combined, ok := transform.(compositeTransformer); ok && combined.priority == 0 {
			for _, nested := range combined.transforms {
				if nested != nil {
					flat = append(flat, nested)
				}
			}
		} else {
			flat = append(flat, transform)
		}
	}
	if len(flat) == 0 {
		return nil
	}
	slices.SortStableFunc(flat, func(a, b Transformer) int { return cmp.Compare(a.Priority(), b.Priority()) })
	return compositeTransformer{transforms: flat}
}

type compositeTransformer struct {
	transforms []Transformer
	priority   int
}

func (composite compositeTransformer) Apply(path []string, value any) (any, error) {
	for index, transform := range composite.transforms {
		var err error
		value, err = transform.Apply(path, value)
		if err != nil {
			return nil, fmt.Errorf("apply transformer %d at %s: %w", index+1, formatPath(path), err)
		}
		if IsRemove(value) {
			return RemoveValue, nil
		}
	}
	return value, nil
}

func (composite compositeTransformer) Priority() int { return composite.priority }

func (composite compositeTransformer) WithPriority(priority int) Transformer {
	composite.priority = priority
	return composite
}
