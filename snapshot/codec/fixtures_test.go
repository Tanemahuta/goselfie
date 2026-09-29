package codec_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tanemahuta/goselfie/snapshot"
	"github.com/tanemahuta/goselfie/snapshot/codec"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
)

var _ = Describe("Stored codec compatibility fixtures", func() {
	for _, version := range []string{codec.Version1, codec.Version2} {
		for _, kind := range []content.Type{content.Text, content.YAML, content.Binary} {
			It("decodes and reproduces "+version+" "+string(kind), func() {
				fixture, err := os.ReadFile(filepath.Join("__testdata__", version, string(kind)+".ss"))
				Expect(err).NotTo(HaveOccurred())
				expected, err := os.ReadFile(filepath.Join("__testdata__", string(kind)+".payload"))
				Expect(err).NotTo(HaveOccurred())
				// A second record checks that decoding stops at the exact record boundary.
				reader := bytes.NewReader(append(append([]byte{}, fixture...), fixture...))
				for range 2 {
					decoded, err := codec.DecodeSnapshot(reader)
					Expect(err).NotTo(HaveOccurred())
					DeferCleanup(decoded.Evict)
					Expect(decoded.Name()).To(Equal([]string{"fixtures", string(kind)}))
					Expect(decoded.ContentType()).To(Equal(kind))
					actual, err := decoded.Contents()
					Expect(err).NotTo(HaveOccurred())
					Expect(actual).To(Equal(expected))
					taken, ok := decoded.(snapshot.Taken)
					Expect(ok).To(BeTrue())
					Expect(taken.Source().Start).To(BeZero())
					Expect(taken.Source().EndExcl).To(Equal(len(fixture)))
					var encoded bytes.Buffer
					Expect(codec.EncodeSnapshotVersion(version, decoded, &encoded)).To(Succeed())
					Expect(encoded.Bytes()).To(Equal(fixture))
				}
				_, err = codec.DecodeSnapshot(reader)
				Expect(err).To(MatchError(io.EOF))
			})
		}
	}
})
