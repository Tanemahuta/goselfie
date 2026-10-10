package yaml

import (
	"fmt"
	"slices"
	"strings"

	sigsyaml "sigs.k8s.io/yaml"
)

// IgnoreOrder sorts the list at its path by the YAML representation of its
// elements. Values other than lists are left unchanged.
func IgnoreOrder(path string) Transformer {
	return pathTransformer(path, sortByYAML)
}

func sortByYAML(value any) (any, error) {
	list, ok := value.([]any)
	if !ok {
		return value, nil
	}
	type entry struct {
		key   string
		value any
	}
	entries := make([]entry, len(list))
	for index, element := range list {
		data, err := sigsyaml.Marshal(element)
		if err != nil {
			return nil, fmt.Errorf("marshal list element %d: %w", index, err)
		}
		entries[index] = entry{key: string(data), value: element}
	}
	slices.SortStableFunc(entries, func(a, b entry) int { return strings.Compare(a.key, b.key) })
	sorted := make([]any, len(entries))
	for index, item := range entries {
		sorted[index] = item.value
	}
	return sorted, nil
}
