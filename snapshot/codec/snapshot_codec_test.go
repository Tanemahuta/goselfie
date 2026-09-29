package codec_test

import (
	"bytes"
	"errors"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tanemahuta/goselfie/snapshot/codec"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
)

var _ = Describe("Snapshot codec", func() {
	When("the codec package loads", func() {
		It("registers both supported versions", func() {
			Expect(codec.ResolveCodec(codec.Version1)).NotTo(BeNil())
			Expect(codec.ResolveCodec(codec.Version2)).NotTo(BeNil())
		})
	})

	When("the legacy version is encoded", func() {
		var encoded bytes.Buffer
		var decoded testData

		BeforeEach(func() {
			value := testData{name: []string{"legacy"}, contentType: content.Text, contents: []byte("v1")}
			Expect(codec.EncodeSnapshotVersion(codec.Version1, value, &encoded)).To(Succeed())
			decodedSnapshot, err := codec.DecodeSnapshot(bytes.NewReader(encoded.Bytes()))
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(decodedSnapshot.Evict)
			decoded = testDataFromSnapshot(decodedSnapshot)
		})

		It("keeps the original colon-framed record", func() {
			Expect(encoded.String()).To(Equal("╔═:v1:text:2:legacy:═╗v1\n"))
			Expect(decoded.name).To(Equal([]string{"legacy"}))
			Expect(decoded.contents).To(Equal([]byte("v1")))
		})
	})

	When("a stream contains consecutive snapshots", func() {
		values := []testData{
			{name: []string{"first"}, contentType: content.Text, contents: []byte("one")},
			{name: []string{"second"}, contentType: content.Text, contents: []byte("two")},
		}
		var decoded []testData

		BeforeEach(func() {
			var encoded bytes.Buffer
			for _, value := range values {
				Expect(codec.EncodeSnapshot(value, &encoded)).To(Succeed())
			}
			reader := bytes.NewReader(encoded.Bytes())
			for range values {
				current, err := codec.DecodeSnapshot(reader)
				Expect(err).NotTo(HaveOccurred())
				DeferCleanup(current.Evict)
				decoded = append(decoded, testDataFromSnapshot(current))
			}
		})

		It("decodes one complete record per call", func() {
			Expect(decoded).To(Equal(values))
		})
	})

	When("a writer accepts only part of a snapshot", func() {
		var encodeErr error

		BeforeEach(func() {
			encodeErr = codec.EncodeSnapshot(testData{name: []string{"short"}, contentType: content.Text, contents: []byte("value")}, shortWriter{})
		})

		It("reports the short write", func() { Expect(encodeErr).To(MatchError(io.ErrShortWrite)) })
	})

	When("snapshot contents cannot be read", func() {
		var encodeErr error
		failure := errors.New("contents failed")

		BeforeEach(func() { encodeErr = codec.EncodeSnapshot(testData{err: failure}, &bytes.Buffer{}) })

		It("returns the content error", func() { Expect(encodeErr).To(MatchError(failure)) })
	})

	for _, testCase := range []struct {
		name, record, message string
	}{
		{"unsupported version", "╔═ v3:text:0:name ═╗\nbody\n", "unsupported snapshot version"},
		{"long version", "╔═ version_name_longer_than_thirty_two_characters:text:0:name ═╗\nbody\n", "unsupported snapshot version"},
		{"unsupported content type", "╔═ v2:opaque:0:name ═╗\n\n", "unsupported data type"},
		{"invalid content length", "╔═ v2:text:nope:name ═╗\n\n", "invalid content length"},
	} {
		testCase := testCase
		When("a snapshot header has "+testCase.name, func() {
			var decodeErr error

			BeforeEach(func() {
				_, decodeErr = codec.DecodeSnapshot(bytes.NewBufferString(testCase.record))
			})

			It("returns a useful error", func() {
				Expect(decodeErr).To(MatchError(ContainSubstring(testCase.message)))
			})
		})
	}
})
