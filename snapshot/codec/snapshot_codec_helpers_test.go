package codec_test

import (
	"io"

	"github.com/tanemahuta/goselfie/snapshot"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
	"github.com/tanemahuta/goselfie/utils"
)

type shortWriter struct{}

func (shortWriter) Write(value []byte) (int, error) {
	if len(value) == 0 {
		return 0, nil
	}
	return len(value) - 1, nil
}

type testData struct {
	name        []string
	contentType content.Type
	contents    []byte
	err         error
}

func (data testData) Name() []string                     { return data.name }
func (data testData) ContentType() content.Type          { return data.contentType }
func (data testData) Contents() ([]byte, error)          { return data.contents, data.err }
func (testData) Evict() error                            { return nil }
func (testData) Promote(utils.Range[int]) snapshot.Taken { return nil }

func testDataFromSnapshot(data snapshot.Data) testData {
	contents, err := data.Contents()
	if err != nil {
		return testData{err: err}
	}
	return testData{name: data.Name(), contentType: data.ContentType(), contents: contents}
}

var _ io.Writer = shortWriter{}
