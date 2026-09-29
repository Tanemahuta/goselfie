package lens

import "github.com/tanemahuta/goselfie/snapshot/codec/content"

// Telescope composes two typed lenses into a single lens.
func Telescope[I any, J any, O any](lhs Lens[I, J], rhs Lens[J, O]) Lens[I, O] {
	return composedLens[I, J, O]{lhs: lhs, rhs: rhs}
}

type composedLens[I any, J any, O any] struct {
	lhs Lens[I, J]
	rhs Lens[J, O]
}

func (lens composedLens[I, J, O]) Apply(input I, contentType content.Type) (O, content.Type, error) {
	middle, middleContentType, err := lens.lhs.Apply(input, contentType)
	if err != nil {
		var zero O
		return zero, content.Unknown, err
	}
	return lens.rhs.Apply(middle, middleContentType)
}
