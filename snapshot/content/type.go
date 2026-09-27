// Package content defines snapshot content types and their matchers.
package content

// Type describes the representation stored for a snapshot's contents.
type Type string

const (
	// Unknown is used when the representation is not specified or detected.
	Unknown Type = "unknown"
	// Text contains readable text stored verbatim.
	Text Type = "text"
	// Binary contains bytes stored as hexadecimal text.
	Binary Type = "binary"
	// YAML contains YAML stored verbatim.
	YAML Type = "yaml"

	// BinaryBytesPerLine is the number of bytes encoded on one hexadecimal line.
	BinaryBytesPerLine = 27
)
