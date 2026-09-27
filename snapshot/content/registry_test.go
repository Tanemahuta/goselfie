package content

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Content matcher registry", func() {
	When("matchers are registered out of priority order", func() {
		var actual Type

		BeforeEach(func() {
			registry := NewRegistry()
			registry.Register(testMatcher{contentType: Text, priority: 20, matches: true})
			registry.Register(testMatcher{contentType: YAML, priority: 10, matches: true})
			registry.Register(testMatcher{contentType: Binary, priority: 30, matches: true})
			actual = registry.Detect([]byte("value"))
		})

		It("checks matching by ascending priority", func() { Expect(actual).To(Equal(YAML)) })
	})

	When("an empty registry detects content", func() {
		var actual Type

		BeforeEach(func() { actual = NewRegistry().Detect([]byte("value")) })

		It("returns Unknown", func() { Expect(actual).To(Equal(Unknown)) })
	})

	When("a matcher is registered for a content type", func() {
		var encoded string

		BeforeEach(func() {
			registry := NewRegistry()
			registry.Register(testMatcher{contentType: Text, matches: true})
			var err error
			encoded, err = registry.Encoder(Text).Encode([]byte("value"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("uses the matcher as the content encoder", func() { Expect(encoded).To(Equal("value")) })
	})

	When("built-in matchers inspect YAML, text, and binary values", func() {
		var actual []Type

		BeforeEach(func() {
			actual = []Type{
				Matchers.Detect([]byte("name: example\n")),
				Matchers.Detect([]byte("name: first\nname: second\n")),
				Matchers.Detect([]byte{0x00, 0xff}),
			}
		})

		It("prefers YAML and falls back to text or binary", func() {
			Expect(actual).To(Equal([]Type{YAML, Text, Binary}))
		})
	})
})

type testMatcher struct {
	contentType Type
	priority    int
	matches     bool
}

func (matcher testMatcher) ContentType() Type   { return matcher.contentType }
func (matcher testMatcher) Priority() int       { return matcher.priority }
func (matcher testMatcher) Matches([]byte) bool { return matcher.matches }
func (testMatcher) Encode(contents []byte) (string, error) {
	return string(contents), nil
}
