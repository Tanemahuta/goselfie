package codec

import (
	"fmt"
	"io"
	"strings"

	"github.com/tanemahuta/goselfie/snapshot"
)

// DecodeSnapshot reads one snapshot, dispatching to the codec named in its header.
func DecodeSnapshot(input io.Reader) (snapshot.Data, error) {
	version, err := readVersion(input)
	if err != nil {
		return nil, err
	}
	registered := ResolveCodec(version)
	if registered == nil {
		return nil, fmt.Errorf("unsupported snapshot version %q", version)
	}
	return registered.DecodeSnapshot(input)
}

func readVersion(input io.Reader) (string, error) {
	prefix := make([]byte, len(headerPrefix))
	read, err := io.ReadFull(input, prefix)
	if err != nil {
		if read == 0 && err == io.EOF {
			return "", io.EOF
		}
		return "", fmt.Errorf("read snapshot header: %w", err)
	}
	if string(prefix) != headerPrefix {
		return "", fmt.Errorf("read snapshot header: expected prefix %q, got %q", headerPrefix, prefix)
	}

	var delimiter [1]byte
	if _, err := io.ReadFull(input, delimiter[:]); err != nil {
		return "", fmt.Errorf("read snapshot header separator: %w", err)
	}
	if delimiter[0] != fieldSep[0] && delimiter[0] != ' ' {
		return "", fmt.Errorf("read snapshot header separator: unsupported %q", delimiter[0])
	}

	var version strings.Builder
	for {
		var fieldByte [1]byte
		if _, err := io.ReadFull(input, fieldByte[:]); err != nil {
			return "", fmt.Errorf("read snapshot version: %w", err)
		}
		if fieldByte[0] == fieldSep[0] {
			if version.Len() == 0 {
				return "", fmt.Errorf("snapshot version is empty")
			}
			return version.String(), nil
		}
		version.WriteByte(fieldByte[0])
	}
}
