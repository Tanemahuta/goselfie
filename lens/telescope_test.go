package lens

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tanemahuta/goselfie/snapshot/codec/content"
)

var _ = Describe("Telescope", func() {
	When("two typed lenses are composed", func() {
		var result []byte
		var resultType content.Type

		BeforeEach(func() {
			first := testTypedLens[int, string](func(value int) (string, error) { return fmt.Sprint(value), nil })
			second := testTypedLens[string, []byte](func(value string) ([]byte, error) { return []byte(value), nil })
			var err error
			result, resultType, err = Telescope(first, second).Apply(42, content.Unknown)
			Expect(err).NotTo(HaveOccurred())
		})

		It("passes the value and content type through both lenses", func() {
			Expect(result).To(Equal([]byte("42")))
			Expect(resultType).To(Equal(content.Unknown))
		})
	})
})
