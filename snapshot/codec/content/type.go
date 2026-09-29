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

	// BinaryMaxCharsPerLine is the maximum width of an encoded hexadecimal row.
	BinaryMaxCharsPerLine = 78
	// BinaryBytesPerLine is the number of bytes encoded on one hexadecimal line.
	// With inter-byte spaces and no trailing space, this uses 77 characters per row.
	BinaryBytesPerLine = BinaryMaxCharsPerLine / 3
)
