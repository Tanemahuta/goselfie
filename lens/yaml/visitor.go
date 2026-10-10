package yaml

import (
	"fmt"
	"strconv"
)

// Visit traverses value depth-first: children are visited before their parent,
// so transforms at a node see already transformed children. Maps and lists
// emptied by removal are removed as well.
func Visit(value any, transforms ...Transformer) (any, error) {
	combined := Combine(transforms...)
	if combined == nil {
		return value, nil
	}
	return visit(combined, nil, value)
}

func visit(transform Transformer, path []string, value any) (any, error) {
	switch current := value.(type) {
	case map[string]any:
		removed := false
		for key, child := range current {
			next, err := visit(transform, childPath(path, key), child)
			if err != nil {
				return nil, err
			}
			if IsRemove(next) {
				delete(current, key)
				removed = true
			} else {
				current[key] = next
			}
		}
		if removed && len(current) == 0 {
			return RemoveValue, nil
		}
	case []any:
		removed := false
		for index := len(current) - 1; index >= 0; index-- {
			next, err := visit(transform, childPath(path, strconv.Itoa(index)), current[index])
			if err != nil {
				return nil, err
			}
			if IsRemove(next) {
				current = append(current[:index], current[index+1:]...)
				removed = true
			} else {
				current[index] = next
			}
		}
		if removed && len(current) == 0 {
			return RemoveValue, nil
		}
		value = current
	}
	updated, err := transform.Apply(path, value)
	if err != nil {
		return nil, fmt.Errorf("transform YAML at %s: %w", formatPath(path), err)
	}
	return updated, nil
}

func childPath(path []string, segment string) []string {
	return append(append(make([]string, 0, len(path)+1), path...), segment)
}
