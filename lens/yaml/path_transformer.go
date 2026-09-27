package yaml

import (
	"fmt"

	"github.com/tanemahuta/goselfie/utils/path"
)

func pathTransformer(expression string, transformer ValueTransformer) Transformer {
	matcher, err := path.CompilePath(expression)
	if err != nil {
		return NewTransformer(nil, func(any) (any, error) {
			return nil, fmt.Errorf("parse YAML path %q: %w", expression, err)
		})
	}
	return NewTransformer(matcher, transformer)
}
