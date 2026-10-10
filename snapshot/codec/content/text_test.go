package content

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	testfixture "github.com/tanemahuta/goselfie/internal/test"
)

var _ = Describe("Text content", func() {
	When("printable UTF-8 text cannot be parsed as YAML", func() {
		contents := testfixture.ReadTestFile("text/duplicate_keys.yaml")
		var actual Type
		var encoded string

		BeforeEach(func() {
			actual = Matchers.Detect(contents)
			var err error
			encoded, err = Matchers.Encoder(Text).Encode(contents)
			Expect(err).NotTo(HaveOccurred())
		})

		It("matches as text and preserves its bytes", func() {
			Expect(actual).To(Equal(Text))
			Expect(encoded).To(Equal(string(contents)))
		})
	})
})
