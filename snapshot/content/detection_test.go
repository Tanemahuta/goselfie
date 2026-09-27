package content

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Content detection", func() {
	When("built-in matchers inspect common content", func() {
		var actual []Type

		BeforeEach(func() {
			actual = []Type{
				Matchers.Detect([]byte("ordinary text\n")),
				Matchers.Detect([]byte("name: example\n")),
				Matchers.Detect([]byte("name: first\nname: second\n")),
				Matchers.Detect([]byte{0x00, 0xff, 0x01}),
			}
		})

		It("classifies valid YAML, invalid YAML text, and binary bytes", func() {
			Expect(actual).To(Equal([]Type{YAML, YAML, Text, Binary}))
		})
	})

	When("a YAML-only registry inspects a stream", func() {
		var actual []Type

		BeforeEach(func() {
			registry := NewRegistry()
			registry.Register(YAMLContent{})
			actual = []Type{
				registry.Detect([]byte("name: first\n---\nname: second\n")),
				registry.Detect([]byte("just words\n")),
				registry.Detect([]byte("42\n")),
				registry.Detect([]byte("name: first\n---\nname: [broken\n")),
				registry.Detect([]byte("name: first\nname: second\n")),
			}
		})

		It("accepts valid collection and scalar documents but rejects decode errors", func() {
			Expect(actual).To(Equal([]Type{YAML, YAML, YAML, Unknown, Unknown}))
		})
	})

	When("a YAML-only registry inspects an empty stream", func() {
		var actual Type

		BeforeEach(func() {
			registry := NewRegistry()
			registry.Register(YAMLContent{})
			actual = registry.Detect([]byte{})
		})

		It("reports no YAML document", func() { Expect(actual).To(Equal(Unknown)) })
	})
})
