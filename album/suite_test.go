package album

import (
	"bytes"
	"io"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tanemahuta/goselfie/snapshot"
	"github.com/tanemahuta/goselfie/snapshot/content"
	"github.com/tanemahuta/goselfie/utils"
)

func TestAlbum(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Album Suite")
}

type testData struct {
	name        []string
	contentType content.Type
	contents    []byte
	err         error
}

func (data testData) Name() []string { return data.name }
func (data testData) ContentType() content.Type {
	if data.contentType == "" {
		return content.Unknown
	}
	return data.contentType
}
func (data testData) Contents() ([]byte, error) {
	return data.contents, data.err
}
func (data testData) Evict() error { return nil }
func (data testData) Promote(source utils.Range[int]) snapshot.Taken {
	taken, err := snapshot.NewTaken(data.name, data.ContentType(), source, func(writer io.Writer) error {
		_, err := bytes.NewReader(data.contents).WriteTo(writer)
		return err
	})
	if err != nil {
		panic(err)
	}
	return taken.Promote(source)
}
