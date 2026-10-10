// SELFIEWRITE
package integration

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tanemahuta/goselfie"
	"github.com/tanemahuta/goselfie/lens"
)

var _ = Describe("file SELFIEWRITE", func() {
	var actual any

	BeforeEach(func() { actual = loadYAML("file_selfiewrite.yaml") })

	It("writes", func() {
		Expect(actual).To(goselfie.MatchSnapshot(lens.ToYAML()).WithName("value"))
	})
})
