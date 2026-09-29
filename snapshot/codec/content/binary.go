package content

import (
	"fmt"
	"math"
)

const binaryPriority = 2

//nolint:gochecknoinits // Register the binary content matcher when this file's package loads.
func init() { Matchers.Register(BinaryContent{}) }

// BinaryContent matches any bytes and encodes them as hexadecimal text.
type BinaryContent struct {
	// BytesPerLine overrides the default row width for legacy snapshot formats.
	BytesPerLine int
}

// ContentType returns Binary.
func (BinaryContent) ContentType() Type { return Binary }

// Priority makes binary detection the final fallback.
func (BinaryContent) Priority() int { return binaryPriority }

// Matches always succeeds because binary is the final content type fallback.
func (BinaryContent) Matches([]byte) bool { return true }

// Encode returns lowercase, space-separated hexadecimal, wrapped at BinaryBytesPerLine.
func (binary BinaryContent) Encode(contents []byte) (string, error) {
	if len(contents) == 0 {
		return "", nil
	}
	if len(contents) > (math.MaxInt+1)/3 {
		return "", fmt.Errorf("binary content length %d is too large", len(contents))
	}
	bytesPerLine := binary.BytesPerLine
	if bytesPerLine <= 0 {
		bytesPerLine = BinaryBytesPerLine
	}
	const digits = "0123456789abcdef"
	encoded := make([]byte, 0, len(contents)*3-1)
	for index, value := range contents {
		if index > 0 {
			if index%bytesPerLine == 0 {
				encoded = append(encoded, '\n')
			} else {
				encoded = append(encoded, ' ')
			}
		}
		encoded = append(encoded, digits[value>>4], digits[value&0x0f])
	}
	return string(encoded), nil
}
