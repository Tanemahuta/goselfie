package failuremessage

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tanemahuta/goselfie/snapshot/codec/content"
)

var _ = Describe("Provider registry", func() {
	When("a provider is registered for a content type", func() {
		var actual Provider
		provider := testProvider{}

		BeforeEach(func() {
			registry := NewRegistry()
			registry.Register(content.Binary, provider)
			actual = registry.Provider(content.Binary)
		})

		It("returns that provider for message generation", func() {
			Expect(actual).To(Equal(provider))
			Expect(actual.FailureMessage([]byte("actual"), []byte("expected"))).To(Equal("actual:expected"))
			Expect(actual.NegateFailureMessage([]byte("actual"), []byte("expected"))).To(Equal("not:actual:expected"))
		})
	})

	When("the built-in text and YAML providers format mismatches", func() {
		var textMessage, yamlMessage string
		var textNegated, yamlNegated string

		BeforeEach(func() {
			text := Providers.Provider(content.Text)
			yaml := Providers.Provider(content.YAML)
			textMessage = text.FailureMessage([]byte("actual text"), []byte("expected text"))
			yamlMessage = yaml.FailureMessage([]byte("name: actual"), []byte("name: expected"))
			textNegated = text.NegateFailureMessage([]byte("actual text"), []byte("expected text"))
			yamlNegated = yaml.NegateFailureMessage([]byte("name: actual"), []byte("name: expected"))
		})

		It("uses Gomega's expected-value and diff format", func() {
			Expect(textMessage).To(ContainSubstring("Expected"))
			Expect(textMessage).To(ContainSubstring("to match snapshot"))
			Expect(textMessage).To(ContainSubstring("actual text"))
			Expect(textMessage).To(ContainSubstring("expected text"))
			Expect(yamlMessage).To(ContainSubstring("Expected"))
			Expect(yamlMessage).To(ContainSubstring("to match YAML snapshot"))
			Expect(yamlMessage).To(ContainSubstring("name: actual"))
			Expect(yamlMessage).To(ContainSubstring("name: expected"))
			Expect(textNegated).To(ContainSubstring("not to match snapshot"))
			Expect(textNegated).To(ContainSubstring("actual text"))
			Expect(yamlNegated).To(ContainSubstring("not to match YAML snapshot"))
			Expect(yamlNegated).To(ContainSubstring("name: expected"))
		})
	})

	When("an unknown content type is looked up", func() {
		var provider, defaultProvider Provider
		var message, negatedMessage string

		BeforeEach(func() {
			registry := NewRegistry()
			provider = registry.Provider(content.Unknown)
			defaultProvider = registry.DefaultProvider()
			message = provider.FailureMessage([]byte{0x00, 0xff}, []byte{0x00, 0x01})
			negatedMessage = provider.NegateFailureMessage([]byte{0x00, 0xff}, []byte{0x00, 0x01})
		})

		It("uses the binary default and diffs spaced hexadecimal text", func() {
			Expect(provider).To(Equal(defaultProvider))
			Expect(message).To(ContainSubstring("to match binary snapshot"))
			Expect(message).To(ContainSubstring("00 ff"))
			Expect(message).To(ContainSubstring("00 01"))
			Expect(negatedMessage).To(ContainSubstring("not to match binary snapshot"))
			Expect(negatedMessage).To(ContainSubstring("00 ff"))
		})
	})
})

type testProvider struct{}

func (testProvider) FailureMessage(actual []byte, expected []byte) string {
	return string(actual) + ":" + string(expected)
}

func (testProvider) NegateFailureMessage(actual []byte, expected []byte) string {
	return "not:" + string(actual) + ":" + string(expected)
}

var _ = Describe("default provider lookup", func() {
	It("returns the registered default for unknown content", func() {
		Expect(For(content.Unknown)).To(Equal(Providers.DefaultProvider()))
	})
})
