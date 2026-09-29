package lens

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	yaml "github.com/tanemahuta/goselfie/lens/yaml"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
	sigsyaml "sigs.k8s.io/yaml"
)

type yamlFixture struct {
	Metadata map[string]any `json:"metadata" yaml:"metadata"`
	Spec     map[string]any `json:"spec" yaml:"spec"`
}

func plainYAML(data []byte) any {
	var value any
	Expect(sigsyaml.Unmarshal(data, &value)).To(Succeed())
	return value
}

var _ = Describe("YAML lens", func() {
	When("serialized YAML is deserialized into a requested type", func() {
		var decoded yamlFixture
		var contentType content.Type

		BeforeEach(func() {
			input := []byte("metadata:\n  uid: original\nspec:\n  enabled: true\n")
			var err error
			decoded, contentType, err = FromYAML[yamlFixture]().Apply(input, content.YAML)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the typed value without claiming stored YAML content", func() {
			Expect(decoded.Metadata).To(Equal(map[string]any{"uid": "original"}))
			Expect(decoded.Spec).To(HaveKeyWithValue("enabled", true))
			Expect(contentType).To(Equal(content.Unknown))
		})
	})

	When("base and transformed YAML lenses serialize a value", func() {
		var baseData, transformedData []byte
		var baseType, transformedType, deserializerType content.Type

		BeforeEach(func() {
			base := ToYAML()
			transformed := base.Transform(yaml.Remove("metadata.uid"))
			var err error
			baseData, baseType, err = base.Apply(map[string]any{"name": "example"}, content.Unknown)
			Expect(err).NotTo(HaveOccurred())
			transformedData, transformedType, err = transformed.Apply(map[string]any{"name": "example"}, content.Unknown)
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
			input := []byte("metadata: [unterminated")
			_, _, decodeErr = FromYAML[yamlFixture]().Apply(input, content.YAML)
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
			_, _, transformErr = projection.Apply(map[string]any{"name": "example"}, content.Unknown)
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
			projection := ToYAML().Transform(nil)
			var err error
			data, _, err = projection.Apply(map[string]any{"empty": map[string]any{}}, content.Unknown)
			Expect(err).NotTo(HaveOccurred())
		})

		It("leaves the value unchanged", func() {
			Expect(plainYAML(data)).To(Equal(map[string]any{"empty": map[string]any{}}))
		})
	})

	When("transforms with priorities are composed", func() {
		var calls []string

		BeforeEach(func() {
			record := func(name string, priority int) yaml.Transformer {
				return yaml.NewTransformer(nil, func(value any) (any, error) { calls = append(calls, name); return value, nil }).WithPriority(priority)
			}
			base := ToYAML()
			projection := base.Transform(record("late", 10)).Transform(record("early", -10))
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
		input := map[string]any{"value": "secret", "keep": true}

		BeforeEach(func() {
			base := ToYAML()
			masked := base.Transform(yaml.Mask("value"))
			removed := base.Transform(yaml.Remove("value"))
			var err error
			original, _, err = base.Apply(input, content.Unknown)
			Expect(err).NotTo(HaveOccurred())
			maskedData, _, err = masked.Apply(input, content.Unknown)
			Expect(err).NotTo(HaveOccurred())
			removedData, _, err = removed.Apply(input, content.Unknown)
			Expect(err).NotTo(HaveOccurred())
		})

		It("keeps each projection's changes independent", func() {
			Expect(plainYAML(original)).To(Equal(input))
			Expect(plainYAML(maskedData)).To(Equal(map[string]any{"value": "<masked>", "keep": true}))
			Expect(plainYAML(removedData)).To(Equal(map[string]any{"keep": true}))
		})
	})

	When("transforms target nested YAML paths", func() {
		var data []byte

		BeforeEach(func() {
			input := yamlFixture{
				Metadata: map[string]any{"uid": "dynamic", "nullable": nil},
				Spec:     map[string]any{"items": []any{map[string]any{"secret": "one"}, map[string]any{"secret": "two"}}},
			}
			var err error
			data, _, err = ToYAML().Transform(yaml.Remove("metadata.uid")).Transform(yaml.Mask("spec.items.0.secret")).Apply(input, content.Unknown)
			Expect(err).NotTo(HaveOccurred())
		})

		It("applies the matching remove and mask transforms", func() {
			Expect(plainYAML(data)).To(Equal(map[string]any{
				"metadata": map[string]any{"nullable": nil},
				"spec": map[string]any{"items": []any{
					map[string]any{"secret": "<masked>"},
					map[string]any{"secret": "two"},
				}},
			}))
		})
	})

	When("child transforms remove every member of collections", func() {
		var data []byte

		BeforeEach(func() {
			input := map[string]any{"metadata": map[string]any{"uid": "x"}, "items": []any{"only"}, "nil": nil}
			var err error
			data, _, err = ToYAML().Transform(yaml.Remove("metadata.uid")).Transform(yaml.Remove("items.0")).Apply(input, content.Unknown)
			Expect(err).NotTo(HaveOccurred())
		})

		It("prunes empty collections but preserves nil values", func() {
			Expect(plainYAML(data)).To(Equal(map[string]any{"nil": nil}))
		})
	})

	When("transforms target recursive and quoted paths", func() {
		var data []byte

		BeforeEach(func() {
			input := map[string]any{
				"spec":   map[string]any{"nested": map[string]any{"uid": "x"}},
				"labels": map[string]any{"example.com/key": "y"},
			}
			var err error
			data, _, err = ToYAML().Transform(yaml.Mask("spec.**.uid")).Transform(yaml.Remove("labels.'example.com/key'")).Apply(input, content.Unknown)
			Expect(err).NotTo(HaveOccurred())
		})

		It("applies both transforms", func() {
			Expect(plainYAML(data)).To(Equal(map[string]any{"spec": map[string]any{"nested": map[string]any{"uid": "<masked>"}}}))
		})
	})
})
