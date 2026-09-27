package contextcache

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tanemahuta/goselfie"
	"github.com/tanemahuta/goselfie/lens"
)

func TestContextCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Context Cache Suite")
}

var _ = Describe("shared test file", func() {
	It("first spec", func() {
		actual := map[string]any{"name": "first"}
		Expect(actual).To(goselfie.MatchSnapshot(lens.ToYAML()))
	})

	It("second spec", func() {
		actual := map[string]any{"name": "second"}
		Expect(actual).To(goselfie.MatchSnapshot(lens.ToYAML()))
	})
})
