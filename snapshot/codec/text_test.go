package codec_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	testfixture "github.com/tanemahuta/goselfie/internal/test"
	"github.com/tanemahuta/goselfie/snapshot/codec"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
)

var _ = Describe("Text snapshots", func() {
	When("text contents are encoded and decoded", func() {
		value := testData{
			name:        []string{"suite", "name > value"},
			contentType: content.Text,
			contents:    testfixture.ReadTestFile("text/input.txt"),
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

		It("writes a newline after the header and preserves text verbatim", func() {
			Expect(encoded.String()).To(Equal(string(testfixture.ReadTestFile("text/expected.ss"))))
			Expect(decodedType).To(Equal(content.Text))
			Expect(decodedContent).To(Equal(value.contents))
		})
	})

	When("a text snapshot payload is truncated", func() {
		var decodeErr error

		BeforeEach(func() {
			_, decodeErr = codec.DecodeSnapshot(bytes.NewReader(testfixture.ReadTestFile("text/truncated.ss")))
		})

		It("reports the missing content bytes", func() {
			Expect(decodeErr).To(MatchError(ContainSubstring("stream text content")))
		})
	})

	When("a text snapshot is missing its header separator", func() {
		var decodeErr error

		BeforeEach(func() {
			_, decodeErr = codec.DecodeSnapshot(bytes.NewReader(testfixture.ReadTestFile("text/missing_header_separator.ss")))
		})

		It("reports the missing newline", func() {
			Expect(decodeErr).To(MatchError(ContainSubstring("snapshot header separator")))
		})
	})

	When("a text snapshot is missing its record separator", func() {
		var decodeErr error

		BeforeEach(func() {
			_, decodeErr = codec.DecodeSnapshot(bytes.NewReader(testfixture.ReadTestFile("text/missing_record_separator.ss")))
		})

		It("reports the incomplete record", func() {
			Expect(decodeErr).To(MatchError(ContainSubstring("snapshot separator")))
		})
	})
})
