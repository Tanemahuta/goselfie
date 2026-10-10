package goselfie

import (
	"github.com/tanemahuta/goselfie/lens"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
)

type constructorLens[I any] func(I) ([]byte, error)

var _ lens.Lens[string, []byte] = constructorLens[string](nil)

func (projection constructorLens[I]) Apply(value I, _ content.Type) ([]byte, content.Type, error) {
	data, err := projection(value)
	return data, content.Unknown, err
}
