package goselfie

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/tanemahuta/goselfie/lens"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
)

var _ = Describe("Matcher constructors", func() {
	var projection lens.Lens[string, []byte]

	BeforeEach(func() {
		projection = constructorLens[string](func(string) ([]byte, error) { return nil, nil })
	})

	When("MatchSnapshot constructs the default matcher", func() {
		var matcher types.GomegaMatcher

		BeforeEach(func() { matcher = MatchSnapshot(projection) })

		It("returns a matcher", func() { Expect(matcher).NotTo(BeNil()) })
	})

	When("a matcher is configured with a name", func() {
		var matcher types.GomegaMatcher

		BeforeEach(func() { matcher = MatchSnapshot(projection).WithName("response") })

		It("returns the configured matcher", func() { Expect(matcher).NotTo(BeNil()) })
	})

	When("MatchSnapshot_TODO constructs a one-shot update matcher", func() {
		var matcher types.GomegaMatcher

		BeforeEach(func() { matcher = MatchSnapshot_TODO(projection) })

		It("returns a matcher", func() { Expect(matcher).NotTo(BeNil()) })
	})
})

type constructorLens[I any] func(I) ([]byte, error)

var _ lens.Lens[string, []byte] = constructorLens[string](nil)

func (projection constructorLens[I]) Apply(value I, _ content.Type) ([]byte, content.Type, error) {
	data, err := projection(value)
	return data, content.Unknown, err
}
