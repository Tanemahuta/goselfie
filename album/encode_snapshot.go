package album

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/tanemahuta/goselfie/snapshot"
	"github.com/tanemahuta/goselfie/snapshot/content"
)

// EncodeSnapshot writes one snapshot in the album format.
func EncodeSnapshot(data snapshot.Data, output io.Writer) error {
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
	encodedContents, err := content.Matchers.Encoder(contentType).Encode(contents)
	if err != nil {
		return fmt.Errorf("encode %s snapshot contents: %w", contentType, err)
	}
	writer := snapshotWriter{output}
	if err := encodeHeader(data.Name(), contentType, len(contents), writer); err != nil {
		return err
	}
	if err := writer.WriteString(encodedContents); err != nil {
		return err
	}
	return writer.WriteString(snapshotSep)
}

func encodeHeader(name []string, contentType content.Type, contentLength int, writer snapshotWriter) error {
	if err := writer.WriteString(headerStart); err != nil {
		return err
	}
	if err := writer.WriteField("v1"); err != nil {
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
	return writer.WriteString(headerEnd)
}

func encodeSnapshotName(name []string, writer snapshotWriter) error {
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
