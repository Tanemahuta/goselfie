package codec

import (
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/tanemahuta/goselfie/snapshot"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
	"github.com/tanemahuta/goselfie/utils"
)

const (
	namePartSep        = ">"
	namePartSepEscaped = "\\" + namePartSep
	fieldSep           = ":"
	snapshotSep        = "\n"
	headerPrefix       = "╔═"
	headerSuffix       = "═╗"
)

// stdCodec implements the shared snapshot serialization model, parameterized by
// the version, header boundary separator, and separator before the contents.
type stdCodec struct {
	version            string
	headerBraceSep     string
	headerContentSep   string
	binaryBytesPerLine int
}

// EncodeSnapshot writes one snapshot in this codec's configured format.
func (standard *stdCodec) EncodeSnapshot(data snapshot.Data, output io.Writer) error {
	contents, err := data.Contents()
	if err != nil {
		return err
	}
	contentType := data.ContentType()
	if contentType == "" || contentType == content.Unknown {
		contentType = content.Matchers.Detect(contents)
	}
	switch contentType {
	case content.Text, content.YAML, content.Binary:
	default:
		return fmt.Errorf("unsupported snapshot content type %q", contentType)
	}
	encoder := content.Matchers.Encoder(contentType)
	if contentType == content.Binary {
		encoder = content.BinaryContent{BytesPerLine: standard.binaryBytesPerLine}
	}
	encodedContents, err := encoder.Encode(contents)
	if err != nil {
		return fmt.Errorf("encode %s snapshot contents: %w", contentType, err)
	}
	writer := writerDecorator{Writer: output}
	if err := encodeHeader(data.Name(), contentType, len(contents), standard.version, standard.headerStart(), standard.headerEnd(), &writer); err != nil {
		return err
	}
	if err := writer.WriteString(standard.headerContentSep); err != nil {
		return err
	}
	if err := writer.WriteString(encodedContents); err != nil {
		return err
	}
	return writer.WriteString(snapshotSep)
}

// DecodeSnapshot reads one snapshot in this codec's configured format.
func (standard *stdCodec) DecodeSnapshot(input io.Reader) (snapshot.Data, error) {
	reader := readerDecorator{
		Reader: input,
		offset: len(standard.headerStart()) + len(standard.version) + len(fieldSep),
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
	encodedName, err := reader.ReadUntil(standard.headerEnd())
	if err != nil {
		return nil, fmt.Errorf("read snapshot name: %w", err)
	}
	if standard.headerContentSep != "" {
		if err := reader.Consume(standard.headerContentSep); err != nil {
			return nil, fmt.Errorf("read snapshot header separator: %w", err)
		}
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
			if err := decodeBinary(&reader, output, length, standard.binaryBytesPerLine); err != nil {
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

func (standard *stdCodec) headerStart() string { return headerPrefix + standard.headerBraceSep }

func (standard *stdCodec) headerEnd() string { return standard.headerBraceSep + headerSuffix }

func encodeHeader(name []string, contentType content.Type, contentLength int, version, start, end string, writer *writerDecorator) error {
	if err := writer.WriteString(start); err != nil {
		return err
	}
	if err := writer.WriteField(version); err != nil {
		return err
	}
	if err := writer.WriteField(string(contentType)); err != nil {
		return err
	}
	if err := writer.WriteField(strconv.Itoa(contentLength)); err != nil {
		return err
	}
	if err := encodeSnapshotName(name, writer); err != nil {
		return err
	}
	return writer.WriteString(end)
}

func encodeSnapshotName(name []string, writer *writerDecorator) error {
	for index, part := range name {
		if index > 0 {
			if err := writer.WriteString(namePartSep); err != nil {
				return err
			}
		}
		if err := writer.WriteString(strings.ReplaceAll(part, namePartSep, namePartSepEscaped)); err != nil {
			return err
		}
	}
	return nil
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

func decodeBinary(reader *readerDecorator, output io.Writer, length, bytesPerLine int) error {
	for index := range length {
		if index > 0 {
			expected := byte(' ')
			if index%bytesPerLine == 0 {
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
		if err := (&writerDecorator{Writer: output}).WriteBytes(decoded); err != nil {
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

var _ Codec = (*stdCodec)(nil)
