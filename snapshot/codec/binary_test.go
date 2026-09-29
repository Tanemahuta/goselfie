package codec_test

import (
	"bytes"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tanemahuta/goselfie/snapshot/codec"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
)

var _ = Describe("Binary snapshots", func() {
	When("binary contents span multiple full rows and a partial final row", func() {
		contents := make([]byte, content.BinaryBytesPerLine*2+3)
		for index := range contents {
			contents[index] = byte(index)
		}
		var encoded bytes.Buffer
		var decodedContents []byte
		var decodedType content.Type

		BeforeEach(func() {
			value := testData{name: []string{"binary"}, contentType: content.Binary, contents: contents}
			Expect(codec.EncodeSnapshot(value, &encoded)).To(Succeed())
			decoded, err := codec.DecodeSnapshot(bytes.NewReader(encoded.Bytes()))
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(decoded.Evict)
			decodedContents, err = decoded.Contents()
			Expect(err).NotTo(HaveOccurred())
			decodedType = decoded.ContentType()
		})

		It("encodes spaced lowercase hex and decodes every byte", func() {
			firstRow := "00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f 10 11 12 13 14 15 16 17 18 19"
			secondRow := "1a 1b 1c 1d 1e 1f 20 21 22 23 24 25 26 27 28 29 2a 2b 2c 2d 2e 2f 30 31 32 33"
			expected := fmt.Sprintf("╔═ v2:binary:%d:binary ═╗\n%s\n%s\n34 35 36\n", len(contents), firstRow, secondRow)
			Expect(encoded.String()).To(Equal(expected))
			Expect(strings.Split(strings.TrimSuffix(strings.TrimPrefix(encoded.String(), "╔═ v2:binary:55:binary ═╗\n"), "\n"), "\n")).To(HaveLen(3))
			Expect(firstRow).To(HaveLen(77)) // 26 byte pairs plus 25 spaces; no trailing space.
			Expect(secondRow).To(HaveLen(77))
			Expect(len(firstRow)).To(BeNumerically("<=", content.BinaryMaxCharsPerLine))
			Expect(len(secondRow)).To(BeNumerically("<=", content.BinaryMaxCharsPerLine))
			Expect(decodedType).To(Equal(content.Binary))
			Expect(decodedContents).To(Equal(contents))
		})
	})

	When("binary contents have no explicit content type", func() {
		var encoded bytes.Buffer
		var decodedContents []byte
		var decodedType content.Type

		BeforeEach(func() {
			contents := []byte{0x00, 0xff, 0x01}
			Expect(codec.EncodeSnapshot(testData{name: []string{"detected"}, contents: contents}, &encoded)).To(Succeed())
			decoded, err := codec.DecodeSnapshot(bytes.NewReader(encoded.Bytes()))
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(decoded.Evict)
			decodedContents, err = decoded.Contents()
			Expect(err).NotTo(HaveOccurred())
			decodedType = decoded.ContentType()
		})

		It("matches the binary matcher and records binary content", func() {
			Expect(encoded.String()).To(HavePrefix("╔═ v2:binary:3:detected ═╗\n00 ff 01\n"))
			Expect(decodedType).To(Equal(content.Binary))
			Expect(decodedContents).To(Equal([]byte{0x00, 0xff, 0x01}))
		})
	})

	When("an empty binary snapshot is encoded and decoded", func() {
		var encoded bytes.Buffer
		var decodedContents []byte
		var decodedType content.Type

		BeforeEach(func() {
			Expect(codec.EncodeSnapshot(testData{name: []string{"empty"}, contentType: content.Binary}, &encoded)).To(Succeed())
			decoded, err := codec.DecodeSnapshot(bytes.NewReader(encoded.Bytes()))
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(decoded.Evict)
			decodedContents, err = decoded.Contents()
			Expect(err).NotTo(HaveOccurred())
			decodedType = decoded.ContentType()
		})

		It("stores no payload bytes", func() {
			Expect(encoded.String()).To(Equal("╔═ v2:binary:0:empty ═╗\n\n"))
			Expect(decodedType).To(Equal(content.Binary))
			Expect(decodedContents).To(BeEmpty())
		})
	})

	When("binary hexadecimal data is malformed", func() {
		var decodeErr error

		BeforeEach(func() {
			_, decodeErr = codec.DecodeSnapshot(bytes.NewBufferString("╔═ v2:binary:2:name ═╗\n00 zz\n"))
		})

		It("reports the invalid byte", func() {
			Expect(decodeErr).To(MatchError(ContainSubstring("decode binary byte")))
		})
	})
})
