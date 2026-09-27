package album

import (
	"bytes"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tanemahuta/goselfie/snapshot"
	"github.com/tanemahuta/goselfie/snapshot/content"
)

var _ = Describe("Snapshot codec", func() {
	When("text data is encoded", func() {
		var encoded bytes.Buffer

		BeforeEach(func() {
			value := testData{
				name:        []string{"suite", "name > value"},
				contentType: content.Text,
				contents:    []byte("first\nvalue"),
			}
			Expect(EncodeSnapshot(value, &encoded)).To(Succeed())
		})

		It("writes the contents verbatim", func() {
			Expect(encoded.String()).To(Equal("╔═:v1:text:11:suite>name \\> value:═╗first\nvalue\n"))
		})
	})

	When("binary data is encoded", func() {
		var encoded bytes.Buffer

		BeforeEach(func() {
			contents := make([]byte, 28)
			for index := range contents {
				contents[index] = byte(index)
			}
			Expect(EncodeSnapshot(testData{name: []string{"binary"}, contents: contents}, &encoded)).To(Succeed())
		})

		It("writes wrapped hexadecimal contents", func() {
			Expect(encoded.String()).To(Equal("╔═:v1:binary:28:binary:═╗" +
				"00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f 10 11 12 13 14 15 16 17 18 19 1a\n" +
				"1b\n"))
		})
	})

	When("a YAML snapshot is encoded and decoded", func() {
		var value testData
		var encoded bytes.Buffer
		var decoded snapshot.Taken

		BeforeEach(func() {
			value = testData{name: []string{"yaml"}, contentType: content.YAML, contents: []byte("name: example\n")}
			Expect(EncodeSnapshot(value, &encoded)).To(Succeed())
			var err error
			decoded, err = DecodeSnapshot(bytes.NewReader(encoded.Bytes()))
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(decoded.Evict)
		})

		It("preserves the YAML header, content type, and bytes", func() {
			Expect(encoded.String()).To(Equal("╔═:v1:yaml:14:yaml:═╗name: example\n\n"))
			Expect(decoded.ContentType()).To(Equal(content.YAML))
			Expect(decoded.Contents()).To(Equal(value.contents))
		})
	})

	When("an empty binary snapshot is encoded", func() {
		var encoded bytes.Buffer

		BeforeEach(func() {
			value := testData{name: []string{"empty"}, contentType: content.Binary}
			Expect(EncodeSnapshot(value, &encoded)).To(Succeed())
		})

		It("records zero bytes without a payload", func() {
			Expect(encoded.String()).To(Equal("╔═:v1:binary:0:empty:═╗\n"))
		})
	})

	When("snapshot contents cannot be read", func() {
		var encodeErr error
		failure := errors.New("contents failed")

		BeforeEach(func() { encodeErr = EncodeSnapshot(testData{err: failure}, &bytes.Buffer{}) })

		It("returns the content error", func() { Expect(encodeErr).To(MatchError(failure)) })
	})

	When("a binary reader writes only part of a decoded byte", func() {
		var decodeErr error

		BeforeEach(func() {
			reader := &snapshotReader{Reader: bytes.NewBufferString("ff")}
			decodeErr = decodeBinary(reader, shortWriter{}, 1)
		})

		It("reports the short write", func() { Expect(decodeErr).To(MatchError(errors.New("short write"))) })
	})

	When("a reader contains consecutive snapshots", func() {
		values := []testData{
			{name: []string{"suite", "text"}, contents: []byte("value")},
			{name: []string{"suite", "binary"}, contents: []byte{0, 1, 0xff}},
		}
		var decoded []snapshot.Taken

		BeforeEach(func() {
			var encoded bytes.Buffer
			for _, value := range values {
				Expect(EncodeSnapshot(value, &encoded)).To(Succeed())
			}
			for range values {
				actual, err := DecodeSnapshot(&encoded)
				Expect(err).NotTo(HaveOccurred())
				decoded = append(decoded, actual)
			}
			DeferCleanup(func() { evictSnapshots(decoded) })
		})

		It("decodes only the next snapshot on each call", func() {
			for index, expected := range values {
				Expect(decoded[index].Name()).To(Equal(expected.Name()))
				Expect(decoded[index].Contents()).To(Equal(expected.contents))
			}
		})
	})

	for _, testCase := range []struct {
		name, contents, message string
	}{
		{"unsupported version", "╔═:v2:text:0:name:═╗\n", "unsupported snapshot version"},
		{"unsupported data type", "╔═:v1:opaque:0:name:═╗\n", "unsupported data type"},
		{"invalid length", "╔═:v1:text:nope:name:═╗\n", "invalid content length"},
		{"truncated contents", "╔═:v1:text:10:name:═╗short\n", "stream text content"},
		{"invalid binary contents", "╔═:v1:binary:2:name:═╗00 zz\n", "decode binary byte"},
		{"missing separator", "╔═:v1:text:5:name:═╗value", "snapshot separator"},
	} {
		testCase := testCase
		When("the encoded snapshot has "+testCase.name, func() {
			var decodeErr error

			BeforeEach(func() {
				_, decodeErr = DecodeSnapshot(bytes.NewBufferString(testCase.contents))
			})

			It("reports the malformed input", func() {
				Expect(decodeErr).To(MatchError(ContainSubstring(testCase.message)))
			})
		})
	}

	When("individual text and binary snapshots are round-tripped", func() {
		values := []snapshot.Data{
			testData{name: []string{"text"}, contents: []byte("multiline\nvalue")},
			testData{name: []string{"binary"}, contents: append([]byte{0, 1, 2, 0xff}, make([]byte, 30)...)},
		}
		var decoded []snapshot.Taken

		BeforeEach(func() {
			var encoded bytes.Buffer
			for _, value := range values {
				Expect(EncodeSnapshot(value, &encoded)).To(Succeed())
				actual, err := DecodeSnapshot(&encoded)
				Expect(err).NotTo(HaveOccurred())
				decoded = append(decoded, actual)
			}
			DeferCleanup(func() { evictSnapshots(decoded) })
		})

		It("preserves each snapshot's name and contents", func() {
			for index, actual := range decoded {
				Expect(actual.Name()).To(Equal(values[index].Name()))
				expectedContents, err := values[index].Contents()
				Expect(err).NotTo(HaveOccurred())
				Expect(actual.Contents()).To(Equal(expectedContents))
			}
		})
	})

	When("a complete album of snapshots is round-tripped", func() {
		values := []snapshot.Data{
			testData{name: []string{"first"}, contents: []byte("text")},
			testData{name: []string{"second"}, contents: []byte{0, 0xff}},
		}
		var decoded []snapshot.Taken

		BeforeEach(func() {
			var encoded bytes.Buffer
			Expect(Encode(values, &encoded)).To(Succeed())
			var err error
			decoded, err = Decode(&encoded)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { evictSnapshots(decoded) })
		})

		It("preserves snapshots in order", func() {
			Expect(decoded).To(HaveLen(2))
			for index, actual := range decoded {
				Expect(actual.Name()).To(Equal(values[index].Name()))
				expectedContents, err := values[index].Contents()
				Expect(err).NotTo(HaveOccurred())
				Expect(actual.Contents()).To(Equal(expectedContents))
			}
		})
	})
})

type shortWriter struct{}

func (shortWriter) Write(value []byte) (int, error) {
	if len(value) == 0 {
		return 0, nil
	}
	return len(value) - 1, nil
}
