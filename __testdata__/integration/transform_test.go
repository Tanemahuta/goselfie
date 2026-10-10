package integration

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tanemahuta/goselfie"
)

var _ = Describe("transform", func() {
	It("stores the original order", func() {
		Expect(loadYAML("transform/original.yaml")).To(goselfie.MatchSnapshot(normalized()).WithName("value"))
	})

	It("stores a shuffled order", func() {
		Expect(loadYAML("transform/shuffled.yaml")).To(goselfie.MatchSnapshot(normalized()).WithName("value"))
	})
})
