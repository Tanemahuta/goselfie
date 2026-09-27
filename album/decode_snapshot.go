package album

import (
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"strconv"

	"github.com/tanemahuta/goselfie/snapshot"
	"github.com/tanemahuta/goselfie/snapshot/content"
	"github.com/tanemahuta/goselfie/utils"
)

// DecodeSnapshot reads one snapshot from an album stream.
func DecodeSnapshot(input io.Reader) (snapshot.Taken, error) {
	reader := snapshotReader{Reader: input}
	if err := reader.Consume(headerStart); err != nil {
		return nil, fmt.Errorf("read snapshot header: %w", err)
	}
	version, err := reader.ReadField()
	if err != nil {
		return nil, fmt.Errorf("read version: %w", err)
	}
	if version != "v1" {
		return nil, fmt.Errorf("unsupported snapshot version %q", version)
	}
	contentTypeField, err := reader.ReadField()
	if err != nil {
		return nil, fmt.Errorf("read data type: %w", err)
	}
	contentType := content.Type(contentTypeField)
	if contentType != content.Unknown && contentType != content.Text && contentType != content.Binary && contentType != content.YAML {
		return nil, fmt.Errorf("unsupported data type %q", contentType)
	}
	lengthField, err := reader.ReadField()
	if err != nil {
		return nil, fmt.Errorf("read content length: %w", err)
	}
	length, err := strconv.Atoi(lengthField)
	if err != nil || length < 0 {
		return nil, fmt.Errorf("invalid content length %q", lengthField)
	}
	encodedName, err := reader.ReadUntil(headerEnd)
	if err != nil {
		return nil, fmt.Errorf("read snapshot name: %w", err)
	}
	encodedLength, err := encodedContentLength(contentType, length)
	if err != nil {
		return nil, err
	}
	source := utils.Range[int]{Start: 0, EndExcl: reader.Offset() + encodedLength + len(snapshotSep)}
	taken, err := snapshot.NewTaken(decodeSnapshotName(encodedName), contentType, source, func(output io.Writer) error {
		switch contentType {
		case content.Unknown, content.Text, content.YAML:
			if err := reader.CopyBytes(output, int64(length)); err != nil {
				return fmt.Errorf("stream text content: %w", err)
			}
		case content.Binary:
			if err := decodeBinary(&reader, output, length); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := reader.Consume(snapshotSep); err != nil {
		_ = taken.Evict()
		return nil, fmt.Errorf("read snapshot separator after %d content bytes: %w", length, err)
	}
	return taken, nil
}

func encodedContentLength(contentType content.Type, length int) (int, error) {
	if contentType != content.Binary || length == 0 {
		return length, nil
	}
	if length > (math.MaxInt+1)/3 {
		return 0, fmt.Errorf("binary content length %d is too large", length)
	}
	return length*3 - 1, nil
}

func decodeBinary(reader *snapshotReader, output io.Writer, length int) error {
	for index := range length {
		if index > 0 {
			expected := byte(' ')
			if index%content.BinaryBytesPerLine == 0 {
				expected = '\n'
			}
			separator, err := reader.ReadBytes(1)
			if err != nil {
				return fmt.Errorf("read binary separator at byte %d: %w", index, err)
			}
			if separator[0] != expected {
				return fmt.Errorf("invalid binary separator at byte %d", index)
			}
		}
		encoded, err := reader.ReadBytes(2)
		if err != nil {
			return fmt.Errorf("read binary byte %d: %w", index, err)
		}
		decoded := []byte{0}
		if _, err := hex.Decode(decoded, encoded); err != nil {
			return fmt.Errorf("decode binary byte %d: %w", index, err)
		}
		if err := (&snapshotWriter{output}).WriteBytes(decoded); err != nil {
			return fmt.Errorf("write binary byte %d: %w", index, err)
		}
	}
	return nil
}

func decodeSnapshotName(encoded string) []string {
	if encoded == "" {
		return nil
	}
	parts := []string{""}
	for index := 0; index < len(encoded); index++ {
		switch {
		case encoded[index] == '\\' && index+1 < len(encoded) && encoded[index+1] == namePartSep[0]:
			parts[len(parts)-1] += namePartSep
			index++
		case encoded[index] == namePartSep[0]:
			parts = append(parts, "")
		default:
			parts[len(parts)-1] += string(encoded[index])
		}
	}
	return parts
}
