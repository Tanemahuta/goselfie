package content

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Binary content", func() {
	When("binary bytes are encoded", func() {
		var actual Type
		var encoded string

		BeforeEach(func() {
			contents := []byte{0x00, 0xff, 0x01}
			actual = Matchers.Detect(contents)
			var err error
			encoded, err = Matchers.Encoder(Binary).Encode(contents)
			Expect(err).NotTo(HaveOccurred())
		})

		It("matches binary and encodes lowercase hex with spaces between bytes", func() {
			Expect(actual).To(Equal(Binary))
			Expect(encoded).To(Equal("00 ff 01"))
		})
	})

	When("binary data spans two full rows and a partial last row", func() {
		var encoded string

		BeforeEach(func() {
			contents := make([]byte, BinaryBytesPerLine*2+3)
			for index := range contents {
				contents[index] = byte(index)
			}
			var err error
			encoded, err = Matchers.Encoder(Binary).Encode(contents)
			Expect(err).NotTo(HaveOccurred())
		})

		It("wraps rows at the configured width without trailing spaces", func() {
			rows := strings.Split(encoded, "\n")
			Expect(rows).To(Equal([]string{
				"00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f 10 11 12 13 14 15 16 17 18 19",
				"1a 1b 1c 1d 1e 1f 20 21 22 23 24 25 26 27 28 29 2a 2b 2c 2d 2e 2f 30 31 32 33",
				"34 35 36",
			}))
			Expect(rows[0]).To(HaveLen(77))
			Expect(rows[1]).To(HaveLen(77))
			for _, row := range rows {
				Expect(len(row)).To(BeNumerically("<=", BinaryMaxCharsPerLine))
				Expect(row).NotTo(HaveSuffix(" "))
			}
		})
	})

	When("no binary encoder is registered for a type", func() {
		var encoded string

		BeforeEach(func() {
			var err error
			encoded, err = NewRegistry().Encoder(Unknown).Encode([]byte{0x10, 0x2a})
			Expect(err).NotTo(HaveOccurred())
		})

		It("uses hexadecimal encoding as the fallback", func() {
			Expect(encoded).To(Equal("10 2a"))
		})
	})
})
