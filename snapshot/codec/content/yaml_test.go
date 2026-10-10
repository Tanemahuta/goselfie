package content

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	testfixture "github.com/tanemahuta/goselfie/internal/test"
)

var _ = Describe("YAML content", func() {
	When("the stream contains valid YAML documents", func() {
		var actual []Type
		var encoded string
		contents := testfixture.ReadTestFile("yaml/multiple_documents.yaml")

		BeforeEach(func() {
			registry := NewRegistry()
			registry.Register(YAMLContent{})
			actual = []Type{
				Matchers.Detect(testfixture.ReadTestFile("yaml/document.yaml")),
				registry.Detect(contents),
				registry.Detect(testfixture.ReadTestFile("yaml/string_scalar.yaml")),
				registry.Detect(testfixture.ReadTestFile("yaml/numeric_scalar.yaml")),
			}
			var err error
			encoded, err = Matchers.Encoder(YAML).Encode(contents)
			Expect(err).NotTo(HaveOccurred())
		})

		It("matches documents and scalars and preserves YAML verbatim", func() {
			Expect(actual).To(Equal([]Type{YAML, YAML, YAML, YAML}))
			Expect(encoded).To(Equal(string(contents)))
		})
	})

	When("a YAML stream contains a decoder error or no document", func() {
		var actual []Type

		BeforeEach(func() {
			registry := NewRegistry()
			registry.Register(YAMLContent{})
			actual = []Type{
				registry.Detect(testfixture.ReadTestFile("yaml/invalid_stream.yaml")),
				registry.Detect(testfixture.ReadTestFile("yaml/invalid_document.yaml")),
				registry.Detect(nil),
			}
		})

		It("returns Unknown", func() { Expect(actual).To(Equal([]Type{Unknown, Unknown, Unknown})) })
	})
})
