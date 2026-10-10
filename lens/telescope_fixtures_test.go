package lens

import "github.com/tanemahuta/goselfie/snapshot/codec/content"

type testTypedLens[I any, O any] func(I) (O, error)

func (lens testTypedLens[I, O]) Apply(input I, _ content.Type) (O, content.Type, error) {
	output, err := lens(input)
	return output, content.Unknown, err
}
