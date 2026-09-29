package content

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("YAML content", func() {
	When("the stream contains valid YAML documents", func() {
		var actual []Type
		var encoded string
		contents := []byte("name: first\n---\nname: second\n")

		BeforeEach(func() {
			registry := NewRegistry()
			registry.Register(YAMLContent{})
			actual = []Type{
				Matchers.Detect([]byte("name: example\n")),
				registry.Detect(contents),
				registry.Detect([]byte("just words\n")),
				registry.Detect([]byte("42\n")),
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
				registry.Detect([]byte("name: first\n---\nname: [broken\n")),
				registry.Detect([]byte("name: [broken\n")),
				registry.Detect(nil),
			}
		})

		It("returns Unknown", func() { Expect(actual).To(Equal([]Type{Unknown, Unknown, Unknown})) })
	})
})
