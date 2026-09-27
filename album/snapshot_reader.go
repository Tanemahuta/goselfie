package album

import (
	"bytes"
	"fmt"
	"io"
)

type snapshotReader struct {
	io.Reader
	offset int
}

func (reader *snapshotReader) Offset() int { return reader.offset }

func (reader *snapshotReader) Consume(expected string) error {
	actual, err := reader.ReadBytes(len(expected))
	if err != nil {
		return err
	}
	if !bytes.Equal(actual, []byte(expected)) {
		return fmt.Errorf("expected %q, got %q", expected, actual)
	}
	return nil
}

func (reader *snapshotReader) ReadField() (string, error) { return reader.ReadUntil(fieldSep) }

func (reader *snapshotReader) ReadUntil(terminator string) (string, error) {
	if terminator == "" {
		return "", fmt.Errorf("terminator must not be empty")
	}
	value := make([]byte, 0)
	for {
		next, err := reader.ReadBytes(1)
		if err != nil {
			return "", fmt.Errorf("value is not terminated by %q: %w", terminator, err)
		}
		value = append(value, next[0])
		if bytes.HasSuffix(value, []byte(terminator)) {
			return string(value[:len(value)-len(terminator)]), nil
		}
	}
}

func (reader *snapshotReader) ReadBytes(length int) ([]byte, error) {
	value := make([]byte, length)
	read, err := io.ReadFull(reader.Reader, value)
	reader.offset += read
	return value, err
}

func (reader *snapshotReader) CopyBytes(writer io.Writer, length int64) error {
	written, err := io.CopyN(writer, reader.Reader, length)
	reader.offset += int(written)
	return err
}
