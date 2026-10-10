package album

import (
	"bytes"
	"io"

	. "github.com/onsi/gomega"
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

type failingTaken struct{ testData }

func (taken failingTaken) Source() utils.Range[int] { return utils.Range[int]{} }
func (taken failingTaken) Promoted() bool           { return false }

func newTaken(name []string, contents string) snapshot.Taken {
	taken, err := snapshot.NewTaken(name, content.Unknown, utils.Range[int]{}, func(writer io.Writer) error {
		_, err := io.WriteString(writer, contents)
		return err
	})
	Expect(err).NotTo(HaveOccurred())
	return taken
}
