package codec

import "io"

type writerDecorator struct{ io.Writer }

func (writer *writerDecorator) WriteBytes(data []byte) error {
	written, err := writer.Write(data)
	if err == nil && written != len(data) {
		return io.ErrShortWrite
	}
	return err
}

func (writer *writerDecorator) WriteField(value string) error {
	if err := writer.WriteString(value); err != nil {
		return err
	}
	return writer.WriteString(fieldSep)
}

func (writer *writerDecorator) WriteString(value string) error {
	return writer.WriteBytes([]byte(value))
}
