package failuremessage

import (
	"encoding/hex"
	"strings"

	"github.com/onsi/gomega/format"
	"github.com/tanemahuta/goselfie/snapshot/content"
)

//nolint:gochecknoinits // Register this provider when its implementation is imported.
func init() {
	Providers.Register(content.Binary, binaryProvider{})
}

// binaryProvider renders bytes as hexadecimal before using Gomega's string diff.
type binaryProvider struct{}

func (binaryProvider) FailureMessage(actual []byte, expected []byte) string {
	return format.MessageWithDiff(formatHex(actual), "to match binary snapshot", formatHex(expected))
}

func (binaryProvider) NegateFailureMessage(actual []byte, expected []byte) string {
	return format.MessageWithDiff(formatHex(actual), "not to match binary snapshot", formatHex(expected))
}

func formatHex(contents []byte) string {
	encoded := hex.EncodeToString(contents)
	if encoded == "" {
		return ""
	}

	const bytesPerLine = 16
	const hexCharsPerByte = 2
	lines := make([]string, 0, (len(contents)+bytesPerLine-1)/bytesPerLine)
	for start := 0; start < len(encoded); {
		end := start + bytesPerLine*hexCharsPerByte
		if end > len(encoded) {
			end = len(encoded)
		}
		line := encoded[start:end]
		var spaced strings.Builder
		spaced.Grow(len(line) + len(line)/2)
		for index := 0; index < len(line); index += hexCharsPerByte {
			if index > 0 {
				spaced.WriteByte(' ')
			}
			spaced.WriteString(line[index : index+hexCharsPerByte])
		}
		lines = append(lines, spaced.String())
		start = end
	}
	return strings.Join(lines, "\n")
}
