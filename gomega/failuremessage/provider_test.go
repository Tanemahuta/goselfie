package failuremessage

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	testfixture "github.com/tanemahuta/goselfie/internal/test"

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
			Expect(actual.FailureMessage(testfixture.ReadTestFile("registered/actual.txt"), testfixture.ReadTestFile("registered/expected.txt"))).To(Equal(string(testfixture.ReadTestFile("registered/expected_message.txt"))))
			Expect(actual.NegateFailureMessage(testfixture.ReadTestFile("registered/actual.txt"), testfixture.ReadTestFile("registered/expected.txt"))).To(Equal(string(testfixture.ReadTestFile("registered/expected_negated_message.txt"))))
		})
	})

	When("the built-in text and YAML providers format mismatches", func() {
		var textMessage, yamlMessage string
		var textNegated, yamlNegated string

		BeforeEach(func() {
			text := Providers.Provider(content.Text)
			yaml := Providers.Provider(content.YAML)
			textMessage = text.FailureMessage(testfixture.ReadTestFile("text/actual.txt"), testfixture.ReadTestFile("text/expected.txt"))
			yamlMessage = yaml.FailureMessage(testfixture.ReadTestFile("yaml/actual.yaml"), testfixture.ReadTestFile("yaml/expected.yaml"))
			textNegated = text.NegateFailureMessage(testfixture.ReadTestFile("text/actual.txt"), testfixture.ReadTestFile("text/expected.txt"))
			yamlNegated = yaml.NegateFailureMessage(testfixture.ReadTestFile("yaml/actual.yaml"), testfixture.ReadTestFile("yaml/expected.yaml"))
		})

		It("uses Gomega's expected-value and diff format", func() {
			Expect(textMessage).To(ContainSubstring("Expected"))
			Expect(textMessage).To(ContainSubstring("to match snapshot"))
			Expect(textMessage).To(ContainSubstring(string(testfixture.ReadTestFile("text/actual.txt"))))
			Expect(textMessage).To(ContainSubstring(string(testfixture.ReadTestFile("text/expected.txt"))))
			Expect(yamlMessage).To(ContainSubstring("Expected"))
			Expect(yamlMessage).To(ContainSubstring("to match YAML snapshot"))
			Expect(yamlMessage).To(ContainSubstring(string(testfixture.ReadTestFile("yaml/actual.yaml"))))
			Expect(yamlMessage).To(ContainSubstring(string(testfixture.ReadTestFile("yaml/expected.yaml"))))
			Expect(textNegated).To(ContainSubstring("not to match snapshot"))
			Expect(textNegated).To(ContainSubstring(string(testfixture.ReadTestFile("text/actual.txt"))))
			Expect(yamlNegated).To(ContainSubstring("not to match YAML snapshot"))
			Expect(yamlNegated).To(ContainSubstring(string(testfixture.ReadTestFile("yaml/expected.yaml"))))
		})
	})

	When("an unknown content type is looked up", func() {
		var provider, defaultProvider Provider
		var message, negatedMessage string

		BeforeEach(func() {
			registry := NewRegistry()
			provider = registry.Provider(content.Unknown)
			defaultProvider = registry.DefaultProvider()
			message = provider.FailureMessage(testfixture.ReadTestFile("binary/actual.bin"), testfixture.ReadTestFile("binary/expected.bin"))
			negatedMessage = provider.NegateFailureMessage(testfixture.ReadTestFile("binary/actual.bin"), testfixture.ReadTestFile("binary/expected.bin"))
		})

		It("uses the binary default and diffs spaced hexadecimal text", func() {
			Expect(provider).To(Equal(defaultProvider))
			Expect(message).To(ContainSubstring("to match binary snapshot"))
			Expect(message).To(ContainSubstring(string(testfixture.ReadTestFile("binary/actual_hex.txt"))))
			Expect(message).To(ContainSubstring(string(testfixture.ReadTestFile("binary/expected_hex.txt"))))
			Expect(negatedMessage).To(ContainSubstring("not to match binary snapshot"))
			Expect(negatedMessage).To(ContainSubstring(string(testfixture.ReadTestFile("binary/actual_hex.txt"))))
		})
	})
})

var _ = Describe("default provider lookup", func() {
	It("returns the registered default for unknown content", func() {
		Expect(For(content.Unknown)).To(Equal(Providers.DefaultProvider()))
	})
})
