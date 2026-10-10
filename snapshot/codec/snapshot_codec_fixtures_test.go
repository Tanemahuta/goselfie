package codec_test

import (
	"io"

	"github.com/tanemahuta/goselfie/snapshot"
)

type shortWriter struct{}

func (shortWriter) Write(value []byte) (int, error) {
	if len(value) == 0 {
		return 0, nil
	}
	return len(value) - 1, nil
}

func testDataFromSnapshot(data snapshot.Data) testData {
	contents, err := data.Contents()
	if err != nil {
		return testData{err: err}
	}
	return testData{name: data.Name(), contentType: data.ContentType(), contents: contents}
}

var _ io.Writer = shortWriter{}
