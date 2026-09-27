package path

import (
	"fmt"
	"strings"
)

type segment struct {
	value string
	glob  bool
}

// Matcher reports whether a complete path matches a compiled expression.
type Matcher func(path []string) bool

// CompilePath parses an expression once for reuse during traversal.
func CompilePath(expression string) (Matcher, error) {
	pattern, err := parsePath(expression)
	if err != nil {
		return nil, err
	}
	return func(path []string) bool { return matchSegments(pattern, path) }, nil
}

func matchSegments(pattern []segment, path []string) bool {
	if len(pattern) == 0 {
		return len(path) == 0
	}
	if pattern[0].glob && pattern[0].value == "**" {
		return matchSegments(pattern[1:], path) || len(path) > 0 && matchSegments(pattern, path[1:])
	}
	if len(path) == 0 {
		return false
	}
	if pattern[0].glob && pattern[0].value == "*" || pattern[0].value == path[0] {
		return matchSegments(pattern[1:], path[1:])
	}
	return false
}

func parsePath(path string) ([]segment, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}
	var result []segment
	for i := 0; i < len(path); {
		if path[i] == '.' {
			return nil, fmt.Errorf("invalid empty segment in path %q", path)
		}
		if path[i] == '\'' || path[i] == '"' {
			quote := path[i]
			start := i
			i++
			var value strings.Builder
			for i < len(path) && path[i] != quote {
				if path[i] == '\\' && i+1 < len(path) {
					i++
				}
				value.WriteByte(path[i])
				i++
			}
			if i == len(path) {
				return nil, fmt.Errorf("unterminated quote at position %d in path %q", start, path)
			}
			i++
			key := value.String()
			result = append(result, segment{value: key, glob: key == "*" || key == "**"})
		} else {
			start := i
			for i < len(path) && path[i] != '.' {
				i++
			}
			value := path[start:i]
			result = append(result, segment{value: value, glob: value == "*" || value == "**"})
		}
		if i < len(path) {
			if path[i] != '.' || i+1 == len(path) {
				return nil, fmt.Errorf("invalid path %q", path)
			}
			i++
		}
	}
	return result, nil
}
