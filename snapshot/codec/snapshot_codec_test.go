package codec_test

import (
	"bytes"
	"errors"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	testfixture "github.com/tanemahuta/goselfie/internal/test"
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
			value := testData{name: []string{"legacy"}, contentType: content.Text, contents: testfixture.ReadTestFile("legacy/input.txt")}
			Expect(codec.EncodeSnapshotVersion(codec.Version1, value, &encoded)).To(Succeed())
			decodedSnapshot, err := codec.DecodeSnapshot(bytes.NewReader(encoded.Bytes()))
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(decodedSnapshot.Evict)
			decoded = testDataFromSnapshot(decodedSnapshot)
		})

		It("keeps the original colon-framed record", func() {
			Expect(encoded.String()).To(Equal(string(testfixture.ReadTestFile("legacy/expected.ss"))))
			Expect(decoded.name).To(Equal([]string{"legacy"}))
			Expect(decoded.contents).To(Equal(testfixture.ReadTestFile("legacy/input.txt")))
		})
	})

	When("a stream contains consecutive snapshots", func() {
		values := []testData{
			{name: []string{"first"}, contentType: content.Text, contents: testfixture.ReadTestFile("consecutive/first.txt")},
			{name: []string{"second"}, contentType: content.Text, contents: testfixture.ReadTestFile("consecutive/second.txt")},
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
			encodeErr = codec.EncodeSnapshot(testData{name: []string{"short"}, contentType: content.Text, contents: testfixture.ReadTestFile("short_write/input.txt")}, shortWriter{})
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
		name, fixture, message string
	}{
		{"unsupported version", "invalid_header/unsupported_version.ss", "unsupported snapshot version"},
		{"long version", "invalid_header/long_version.ss", "unsupported snapshot version"},
		{"unsupported content type", "invalid_header/unsupported_content_type.ss", "unsupported data type"},
		{"invalid content length", "invalid_header/invalid_content_length.ss", "invalid content length"},
	} {
		testCase := testCase
		When("a snapshot header has "+testCase.name, func() {
			var decodeErr error

			BeforeEach(func() {
				_, decodeErr = codec.DecodeSnapshot(bytes.NewReader(testfixture.ReadTestFile(testCase.fixture)))
			})

			It("returns a useful error", func() {
				Expect(decodeErr).To(MatchError(ContainSubstring(testCase.message)))
			})
		})
	}
})
