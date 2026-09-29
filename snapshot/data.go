package snapshot

import (
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
	"github.com/tanemahuta/goselfie/utils"
)

// Data exposes a snapshot's name, content type, bytes, and temporary lifecycle.
type Data interface {
	// Name returns the snapshot name parts, including the container path.
	Name() []string
	// ContentType returns the representation of the snapshot contents.
	ContentType() content.Type
	// Contents returns the snapshot bytes.
	Contents() ([]byte, error)
	// Evict removes the temporary data associated with the snapshot.
	Evict() error
	// Promote marks this stored snapshot as used and records its matcher source.
	Promote(source utils.Range[int]) Taken
}
