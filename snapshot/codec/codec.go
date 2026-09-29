// Package codec provides versioned snapshot codecs.
package codec

import (
	"io"
	"sync"

	"github.com/tanemahuta/goselfie/snapshot"
)

// Codec reads and writes snapshots in a specific format version.
type Codec interface {
	// EncodeSnapshot writes the given snapshot to the provided writer.
	EncodeSnapshot(snapshot snapshot.Data, writer io.Writer) error
	// DecodeSnapshot reads one snapshot from the reader.
	DecodeSnapshot(input io.Reader) (snapshot.Data, error)
}

//nolint:gochecknoglobals // Codecs are registered once and shared by album operations.
var codecRegistry = struct {
	sync.RWMutex
	backing map[string]Codec
}{backing: make(map[string]Codec)}

// RegisterCodec registers a codec for a specific snapshot version.
func RegisterCodec(version string, registered Codec) {
	codecRegistry.Lock()
	defer codecRegistry.Unlock()
	codecRegistry.backing[version] = registered
}

// ResolveCodec retrieves the codec registered for a specific snapshot version.
func ResolveCodec(version string) Codec {
	codecRegistry.RLock()
	defer codecRegistry.RUnlock()
	return codecRegistry.backing[version]
}

const (
	// Version1 is the original colon-framed snapshot format without a body separator.
	Version1 = "v1"
	// Version2 is the snapshot format with spaces inside its header braces and a newline after its header.
	Version2 = "v2"
	// CurrentVersion is the format used when writing new albums.
	CurrentVersion = Version2
)
