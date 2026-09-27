package lens

import "github.com/tanemahuta/goselfie/snapshot/content"

// Lens transforms a snapshot input into another representation.
type Lens[I any, O any] interface {
	// Apply transforms input using its contentType and returns the output content type.
	Apply(input I, contentType content.Type) (O, content.Type, error)
}
