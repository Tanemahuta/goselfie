// Package codec provides versioned snapshot codecs.
package codec

import (
	"fmt"
	"io"

	"github.com/tanemahuta/goselfie/snapshot"
)

// EncodeSnapshot writes a snapshot using the current format version.
func EncodeSnapshot(data snapshot.Data, writer io.Writer) error {
	return EncodeSnapshotVersion(CurrentVersion, data, writer)
}

// EncodeSnapshotVersion writes a snapshot using the requested format version.
func EncodeSnapshotVersion(version string, data snapshot.Data, writer io.Writer) error {
	registered := ResolveCodec(version)
	if registered == nil {
		return fmt.Errorf("no codec registered for snapshot version %q", version)
	}
	return registered.EncodeSnapshot(data, writer)
}
