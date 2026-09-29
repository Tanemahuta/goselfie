package codec_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tanemahuta/goselfie/snapshot/codec"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
)

var _ = Describe("YAML snapshots", func() {
	When("YAML contents are encoded and decoded", func() {
		value := testData{
			name:        []string{"document"},
			contentType: content.YAML,
			contents:    []byte("name: example\n"),
		}
		var encoded bytes.Buffer
		var decodedContent []byte
		var decodedType content.Type

		BeforeEach(func() {
			Expect(codec.EncodeSnapshot(value, &encoded)).To(Succeed())
			decoded, err := codec.DecodeSnapshot(bytes.NewReader(encoded.Bytes()))
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(decoded.Evict)
			decodedContent, err = decoded.Contents()
			Expect(err).NotTo(HaveOccurred())
			decodedType = decoded.ContentType()
		})

		It("preserves the YAML header, content type, and document bytes", func() {
			Expect(encoded.String()).To(Equal("╔═ v2:yaml:14:document ═╗\nname: example\n\n"))
			Expect(decodedType).To(Equal(content.YAML))
			Expect(decodedContent).To(Equal(value.contents))
		})
	})
})
