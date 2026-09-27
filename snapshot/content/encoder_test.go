package content

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("String encoders", func() {
	When("built-in matchers encode their content types", func() {
		var encoded []string

		BeforeEach(func() {
			inputs := []struct {
				contentType Type
				contents    []byte
			}{
				{Text, []byte("hello\n")},
				{YAML, []byte("name: example\n")},
				{Binary, []byte{0x00, 0xff, 0x01}},
			}
			encoded = make([]string, len(inputs))
			for index, input := range inputs {
				var err error
				encoded[index], err = Matchers.Encoder(input.contentType).Encode(input.contents)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("keeps text and YAML verbatim and encodes binary as spaced hexadecimal", func() {
			Expect(encoded).To(Equal([]string{"hello\n", "name: example\n", "00 ff 01"}))
		})
	})

	When("an empty registry encodes an unknown content type", func() {
		var encoded string

		BeforeEach(func() {
			var err error
			encoded, err = NewRegistry().Encoder(Unknown).Encode([]byte{0x10, 0x2a})
			Expect(err).NotTo(HaveOccurred())
		})

		It("uses hexadecimal encoding by default", func() { Expect(encoded).To(Equal("10 2a")) })
	})

	When("binary contents are wider than one hexadecimal line", func() {
		var encoded string

		BeforeEach(func() {
			contents := make([]byte, BinaryBytesPerLine+1)
			var err error
			encoded, err = Matchers.Encoder(Binary).Encode(contents)
			Expect(err).NotTo(HaveOccurred())
		})

		It("starts the next byte on a new line", func() {
			Expect(encoded).To(HavePrefix("00 " + "00 "))
			Expect(encoded).To(ContainSubstring("\n00"))
			Expect(encoded[:strings.IndexByte(encoded, '\n')]).To(HaveLen(80))
		})
	})
})
