package album

import (
	"io"
)

type snapshotWriter struct{ io.Writer }

func (writer *snapshotWriter) WriteBytes(data []byte) error {
	written, err := writer.Write(data)
	if err == nil && written != len(data) {
		return io.ErrShortWrite
	}
	return err
}

func (writer *snapshotWriter) WriteField(value string) error {
	if err := writer.WriteString(value); err != nil {
		return err
	}
	return writer.WriteString(fieldSep)
}

func (writer *snapshotWriter) WriteString(value string) error {
	return writer.WriteBytes([]byte(value))
}
