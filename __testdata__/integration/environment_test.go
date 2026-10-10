package integration

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tanemahuta/goselfie"
	"github.com/tanemahuta/goselfie/lens"
)

var _ = Describe("environment", func() {
	It("updates existing", func() {
		actual := loadYAML("environment/updates_existing.yaml")
		Expect(actual).To(goselfie.MatchSnapshot(lens.ToYAML()).WithName("value"))
	})

	It("writes another container", func() {
		actual := loadYAML("environment/another_container.yaml")
		Expect(actual).To(goselfie.MatchSnapshot(lens.ToYAML()).WithName("value"))
	})
})
