package integration

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tanemahuta/goselfie"
	"github.com/tanemahuta/goselfie/lens"
)

var _ = Describe("environment", func() {
	It("updates existing", func() {
		actual := map[string]any{"name": "environment-original"}
		Expect(actual).To(goselfie.MatchSnapshot(lens.ToYAML()).WithName("value"))
	})

	It("writes another container", func() {
		actual := map[string]any{"name": "environment-second"}
		Expect(actual).To(goselfie.MatchSnapshot(lens.ToYAML()).WithName("value"))
	})
})
