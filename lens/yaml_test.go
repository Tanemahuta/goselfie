package lens

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	testfixture "github.com/tanemahuta/goselfie/internal/test"
	yaml "github.com/tanemahuta/goselfie/lens/yaml"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
)

var _ = Describe("YAML lens", func() {
	When("serialized YAML is deserialized into a requested type", func() {
		var decoded yamlFixture
		var contentType content.Type

		BeforeEach(func() {
			var err error
			decoded, contentType, err = FromYAML[yamlFixture]().Apply(testfixture.ReadTestFile("deserialize/input.yaml"), content.YAML)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the typed value without claiming stored YAML content", func() {
			Expect(decoded).To(Equal(loadYAML[yamlFixture]("deserialize/input.yaml")))
			Expect(decoded.Spec).To(HaveKeyWithValue("enabled", true))
			Expect(contentType).To(Equal(content.Unknown))
		})
	})

	When("base and transformed YAML lenses serialize a value", func() {
		var baseData, transformedData []byte
		var baseType, transformedType, deserializerType content.Type

		BeforeEach(func() {
			input := loadYAML[any]("serialize/input.yaml")
			base := ToYAML()
			transformed := base.Transform(yaml.Remove("metadata.uid"))
			var err error
			baseData, baseType, err = base.Apply(input, content.Unknown)
			Expect(err).NotTo(HaveOccurred())
			transformedData, transformedType, err = transformed.Apply(input, content.Unknown)
			Expect(err).NotTo(HaveOccurred())
			_, deserializerType, err = FromYAML[yamlFixture]().Apply(baseData, baseType)
			Expect(err).NotTo(HaveOccurred())
		})

		It("marks only serialization outputs as YAML", func() {
			Expect(baseData).NotTo(BeEmpty())
			Expect(transformedData).NotTo(BeEmpty())
			Expect(baseType).To(Equal(content.YAML))
			Expect(transformedType).To(Equal(content.YAML))
			Expect(deserializerType).To(Equal(content.Unknown))
		})
	})

	When("malformed YAML is deserialized", func() {
		var decodeErr error

		BeforeEach(func() {
			_, _, decodeErr = FromYAML[yamlFixture]().Apply(testfixture.ReadTestFile("deserialize/malformed.yaml"), content.YAML)
		})

		It("reports the deserialization stage", func() {
			Expect(decodeErr).To(MatchError(ContainSubstring("deserialize YAML")))
		})
	})

	When("serialization or transformation fails", func() {
		var serializationErr, transformErr error
		cause := errors.New("transform failed")

		BeforeEach(func() {
			_, _, serializationErr = ToYAML().Apply(make(chan int), content.Unknown)
			projection := ToYAML().Transform(yaml.NewTransformer(nil, func(any) (any, error) { return nil, cause }))
			_, _, transformErr = projection.Apply(loadYAML[any]("serialize/input.yaml"), content.Unknown)
		})

		It("identifies the failing stage and preserves the cause", func() {
			Expect(serializationErr).To(MatchError(ContainSubstring("serialize value to YAML")))
			Expect(errors.Is(transformErr, cause)).To(BeTrue())
			Expect(transformErr.Error()).To(ContainSubstring("apply YAML transforms"))
		})
	})

	When("a YAML lens receives a nil transform", func() {
		var data []byte

		BeforeEach(func() {
			var err error
			data, _, err = ToYAML().Transform(nil).Apply(loadYAML[any]("nil_transform/input.yaml"), content.Unknown)
			Expect(err).NotTo(HaveOccurred())
		})

		It("leaves the value unchanged", func() {
			Expect(plainYAML(data)).To(Equal(loadYAML[any]("nil_transform/input.yaml")))
		})
	})

	When("transforms with priorities are composed", func() {
		var calls []string

		BeforeEach(func() {
			calls = nil
			base := ToYAML()
			projection := base.Transform(recordingTransformer(&calls, "late", 10)).Transform(recordingTransformer(&calls, "early", -10))
			_, _, err := projection.Apply("value", content.Unknown)
			Expect(err).NotTo(HaveOccurred())
			Expect(calls).To(Equal([]string{"early", "late"}))
			calls = nil
			_, _, err = base.Apply("value", content.Unknown)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not mutate the source lens", func() { Expect(calls).To(BeEmpty()) })
	})

	When("two projections branch from one YAML lens", func() {
		var original, maskedData, removedData []byte

		BeforeEach(func() {
			base := ToYAML()
			var err error
			original, _, err = base.Apply(loadYAML[any]("branches/input.yaml"), content.Unknown)
			Expect(err).NotTo(HaveOccurred())
			maskedData, _, err = base.Transform(yaml.Mask("value")).Apply(loadYAML[any]("branches/input.yaml"), content.Unknown)
			Expect(err).NotTo(HaveOccurred())
			removedData, _, err = base.Transform(yaml.Remove("value")).Apply(loadYAML[any]("branches/input.yaml"), content.Unknown)
			Expect(err).NotTo(HaveOccurred())
		})

		It("keeps each projection's changes independent", func() {
			Expect(plainYAML(original)).To(Equal(loadYAML[any]("branches/input.yaml")))
			Expect(plainYAML(maskedData)).To(Equal(loadYAML[any]("branches/masked.yaml")))
			Expect(plainYAML(removedData)).To(Equal(loadYAML[any]("branches/removed.yaml")))
		})
	})

	When("transforms target nested YAML paths", func() {
		var data []byte

		BeforeEach(func() {
			var err error
			data, _, err = ToYAML().Transform(yaml.Remove("metadata.uid")).Transform(yaml.Mask("spec.items.0.secret")).
				Apply(loadYAML[yamlFixture]("nested_paths/input.yaml"), content.Unknown)
			Expect(err).NotTo(HaveOccurred())
		})

		It("applies the matching remove and mask transforms", func() {
			Expect(plainYAML(data)).To(Equal(loadYAML[any]("nested_paths/expected.yaml")))
		})
	})

	When("child transforms remove every member of collections", func() {
		var data []byte

		BeforeEach(func() {
			var err error
			data, _, err = ToYAML().Transform(yaml.Remove("metadata.uid")).Transform(yaml.Remove("items.0")).
				Apply(loadYAML[any]("emptied_collections/input.yaml"), content.Unknown)
			Expect(err).NotTo(HaveOccurred())
		})

		It("prunes empty collections but preserves nil values", func() {
			Expect(plainYAML(data)).To(Equal(loadYAML[any]("emptied_collections/expected.yaml")))
		})
	})

	When("remove, mask, and ignore-order transforms are combined", func() {
		var data []byte

		BeforeEach(func() {
			transform := yaml.Combine(
				yaml.IgnoreOrder("items"),
				yaml.Mask("items.*.secret"),
				yaml.Remove("**.uid"),
			)
			var err error
			data, _, err = ToYAML().Transform(transform).Apply(loadYAML[any]("combined/input.yaml"), content.Unknown)
			Expect(err).NotTo(HaveOccurred())
		})

		It("transforms children before sorting and leaves other lists in place", func() {
			Expect(data).To(Equal(testfixture.ReadTestFile("combined/expected.yaml")))
		})
	})

	When("transforms target recursive and quoted paths", func() {
		var data []byte

		BeforeEach(func() {
			var err error
			data, _, err = ToYAML().Transform(yaml.Mask("spec.**.uid")).Transform(yaml.Remove("labels.'example.com/key'")).
				Apply(loadYAML[any]("recursive_quoted/input.yaml"), content.Unknown)
			Expect(err).NotTo(HaveOccurred())
		})

		It("applies both transforms", func() {
			Expect(plainYAML(data)).To(Equal(loadYAML[any]("recursive_quoted/expected.yaml")))
		})
	})
})
