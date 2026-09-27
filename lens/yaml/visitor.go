package yaml

import (
	"fmt"
	"reflect"
	"strconv"
)

// Visit applies transforms at each node, then descends if the node was not
// replaced. Maps and lists emptied by removal are removed as well.
func Visit(value any, transforms ...Transformer) (any, error) {
	combined := Combine(transforms...)
	if combined == nil {
		return value, nil
	}
	return visit(combined, nil, value)
}

func visit(transform Transformer, path []string, value any) (any, error) {
	updated, err := transform.Apply(path, value)
	if err != nil {
		return nil, fmt.Errorf("transform YAML at %s: %w", formatPath(path), err)
	}
	if IsRemove(updated) {
		return RemoveValue, nil
	}
	if !sameValue(value, updated) {
		return updated, nil
	}
	switch current := updated.(type) {
	case map[string]any:
		for key, child := range current {
			next, err := visit(transform, append(append([]string(nil), path...), key), child)
			if err != nil {
				return nil, err
			}
			if IsRemove(next) {
				delete(current, key)
			} else {
				current[key] = next
			}
		}
		if len(current) == 0 {
			return RemoveValue, nil
		}
	case []any:
		for index := len(current) - 1; index >= 0; index-- {
			next, err := visit(transform, append(append([]string(nil), path...), strconv.Itoa(index)), current[index])
			if err != nil {
				return nil, err
			}
			if IsRemove(next) {
				current = append(current[:index], current[index+1:]...)
			} else {
				current[index] = next
			}
		}
		if len(current) == 0 {
			return RemoveValue, nil
		}
		updated = current
	}
	return updated, nil
}

func sameValue(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	left, right := reflect.ValueOf(a), reflect.ValueOf(b)
	if left.Type() != right.Type() {
		return false
	}
	switch left.Kind() {
	case reflect.Map, reflect.Slice, reflect.Pointer, reflect.Func, reflect.Chan:
		return left.Pointer() == right.Pointer()
	default:
		return reflect.DeepEqual(a, b)
	}
}
