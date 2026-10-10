package codec_test

import (
	"github.com/tanemahuta/goselfie/snapshot"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
	"github.com/tanemahuta/goselfie/utils"
)

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
