package content

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	testfixture "github.com/tanemahuta/goselfie/internal/test"
)

var _ = Describe("Content matcher registry", func() {
	When("matchers are registered out of priority order", func() {
		var actual Type

		BeforeEach(func() {
			registry := NewRegistry()
			registry.Register(testMatcher{contentType: Text, priority: 20, matches: true})
			registry.Register(testMatcher{contentType: YAML, priority: 10, matches: true})
			registry.Register(testMatcher{contentType: Binary, priority: 30, matches: true})
			actual = registry.Detect(testfixture.ReadTestFile("registry/value.txt"))
		})

		It("checks matching by ascending priority", func() { Expect(actual).To(Equal(YAML)) })
	})

	When("an empty registry detects content", func() {
		var actual Type

		BeforeEach(func() { actual = NewRegistry().Detect(testfixture.ReadTestFile("registry/value.txt")) })

		It("returns Unknown", func() { Expect(actual).To(Equal(Unknown)) })
	})

	When("a matcher is registered for a content type", func() {
		var encoded string

		BeforeEach(func() {
			registry := NewRegistry()
			registry.Register(testMatcher{contentType: Text, matches: true})
			var err error
			encoded, err = registry.Encoder(Text).Encode(testfixture.ReadTestFile("registry/value.txt"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("uses the matcher as the content encoder", func() { Expect(encoded).To(Equal(string(testfixture.ReadTestFile("registry/value.txt")))) })
	})

})
