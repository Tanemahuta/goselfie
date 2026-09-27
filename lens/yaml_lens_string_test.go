package lens

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("YAML lens descriptions", func() {
	It("uses concise names when formatted", func() {
		Expect(fmt.Sprint(ToYAML())).To(Equal("toYAML"))
		Expect(fmt.Sprint(FromYAML[map[string]any]())).To(Equal("fromYAML"))
	})
})
