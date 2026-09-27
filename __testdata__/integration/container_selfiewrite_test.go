package integration

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tanemahuta/goselfie"
	"github.com/tanemahuta/goselfie/lens"
)

var _ = Describe("container SELFIEWRITE", func() {
	var actual map[string]any

	BeforeEach(func() { actual = map[string]any{"name": "container-always"} })

	// SELFIEWRITE
	It("writes", func() {
		Expect(actual).To(goselfie.MatchSnapshot(lens.ToYAML()).WithName("value"))
	})
})
