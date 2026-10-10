// selfieonce
package integration

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tanemahuta/goselfie"
	"github.com/tanemahuta/goselfie/lens"
)

func TestSnapshots(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = Describe("file selfieonce", func() {
	var actual any

	BeforeEach(func() { actual = loadYAML("file_selfieonce.yaml") })

	It("writes", func() {
		Expect(actual).To(goselfie.MatchSnapshot(lens.ToYAML()).WithName("value"))
	})
})
