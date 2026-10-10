package k8s

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tanemahuta/goselfie/lens"
	"github.com/tanemahuta/goselfie/lens/yaml"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
	sigsyaml "sigs.k8s.io/yaml"
)

var _ = Describe("Default Kubernetes transformer", func() {
	When("resources contain runtime metadata", func() {
		var result any

		BeforeEach(func() {
			var err error
			result, err = yaml.Visit(loadYAML[any]("runtime_metadata/input.yaml"), DefaultTransformer())
			Expect(err).NotTo(HaveOccurred())
		})

		It("removes runtime metadata at every resource level", func() {
			Expect(result).To(Equal(loadYAML[any]("runtime_metadata/expected.yaml")))
		})
	})

	When("a typed resource contains only runtime metadata", func() {
		var plain any

		BeforeEach(func() {
			value := loadYAML[resource]("typed_resource/input.yaml")
			data, _, err := lens.ToYAML().Transform(DefaultTransformer()).Apply(value, content.Unknown)
			Expect(err).NotTo(HaveOccurred())
			Expect(sigsyaml.Unmarshal(data, &plain)).To(Succeed())
		})

		It("removes metadata emptied by normalization", func() {
			Expect(plain).To(Equal(loadYAML[any]("typed_resource/expected.yaml")))
		})
	})
})
