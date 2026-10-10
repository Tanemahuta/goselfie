package yaml

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tanemahuta/goselfie/utils/path"
)

var _ = Describe("YAML visitor", func() {
	When("a transformer replaces a parent node", func() {
		var calls []string
		var result any

		BeforeEach(func() {
			calls = nil
			result = visitYAML("replace_parent/input.yaml",
				recordingPathTransformer(&calls, "replace", loadYAML("replace_parent/replacement.yaml")))
		})

		It("keeps the replacement and skips its child", func() {
			Expect(result).To(Equal(loadYAML("replace_parent/expected.yaml")))
			Expect(calls).NotTo(ContainElement("replace.child"))
		})
	})

	When("a transform returns nil or the removal marker", func() {
		var result any

		BeforeEach(func() {
			transform := NewTransformer(
				func(path []string) bool { return formatPath(path) == "nullable" },
				func(any) (any, error) { return nil, nil },
			)
			result = visitYAML("nil_and_remove/input.yaml", transform, Remove("discard"))
		})

		It("preserves nil and removes the marked value", func() {
			Expect(result).To(Equal(loadYAML("nil_and_remove/expected.yaml")))
		})
	})

	When("multiple list elements are removed", func() {
		var result any

		BeforeEach(func() {
			result = visitYAML("remove_list_elements/input.yaml", Remove("items.0"), Remove("items.2"))
		})

		It("removes original indexes without skipping shifted values", func() {
			Expect(result).To(Equal(loadYAML("remove_list_elements/expected.yaml")))
		})
	})

	When("a transform removes the root", func() {
		var result any

		BeforeEach(func() {
			transform := NewTransformer(
				func(path []string) bool { return len(path) == 0 },
				func(any) (any, error) { return RemoveValue, nil },
			)
			result = visitYAML("remove_root/input.yaml", transform)
		})

		It("returns the removal marker", func() { Expect(IsRemove(result)).To(BeTrue()) })
	})

	When("nil transforms appear in nested combinations", func() {
		var unchanged, changed any

		BeforeEach(func() {
			unchanged = visitYAML("nil_transforms/unchanged.yaml", nil, Combine(nil))
			changed = visitYAML("nil_transforms/input.yaml", Combine(nil, Combine(nil, Mask("value"), nil), nil))
		})

		It("ignores nil transforms", func() {
			Expect(Combine(nil, Combine(nil))).To(BeNil())
			Expect(unchanged).To(Equal(loadYAML("nil_transforms/unchanged.yaml")))
			Expect(changed).To(Equal(loadYAML("nil_transforms/expected.yaml")))
		})
	})

	When("flat and nested transforms have different priorities", func() {
		var flatCalls, nestedCalls []string

		BeforeEach(func() {
			flatCalls, nestedCalls = nil, nil
			_, err := Visit("value",
				recordingTransformer(&flatCalls, "last", 10),
				recordingTransformer(&flatCalls, "first tie", 0),
				recordingTransformer(&flatCalls, "first", -10),
				recordingTransformer(&flatCalls, "second tie", 0),
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = Visit("value",
				Combine(recordingTransformer(&nestedCalls, "late", 10), recordingTransformer(&nestedCalls, "early", -10)),
				recordingTransformer(&nestedCalls, "middle", 0),
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("orders by priority and preserves ties", func() {
			Expect(flatCalls).To(Equal([]string{"first", "first tie", "second tie", "last"}))
			Expect(nestedCalls).To(Equal([]string{"early", "middle", "late"}))
		})
	})

	When("a transformer has a path matcher", func() {
		var calls int
		var unchanged, changed any

		BeforeEach(func() {
			calls = 0
			matcher, err := path.CompilePath("spec.*.'example.com/key'")
			Expect(err).NotTo(HaveOccurred())
			transform := NewTransformer(matcher, func(any) (any, error) { calls++; return "changed", nil })
			unchanged, err = transform.Apply([]string{"spec", "item", "other"}, "original")
			Expect(err).NotTo(HaveOccurred())
			changed, err = transform.Apply([]string{"spec", "item", "example.com/key"}, "original")
			Expect(err).NotTo(HaveOccurred())
		})

		It("checks the path before invoking the value transformer", func() {
			Expect(unchanged).To(Equal("original"))
			Expect(changed).To(Equal("changed"))
			Expect(calls).To(Equal(1))
		})
	})

	When("a transform contains an invalid path", func() {
		var applyErr error

		BeforeEach(func() { _, applyErr = Remove("spec..uid").Apply([]string{"spec", "uid"}, "value") })

		It("includes the path expression in the error", func() {
			Expect(applyErr).To(MatchError(ContainSubstring("parse YAML path \"spec..uid\"")))
		})
	})

	When("a transform or composite priority changes", func() {
		var original, prioritized, composite, prioritizedComposite Transformer

		BeforeEach(func() {
			original = Mask("value")
			composite = Combine(original)
			prioritized = original.WithPriority(5)
			prioritizedComposite = composite.WithPriority(7)
		})

		It("returns new wrappers with the requested priorities", func() {
			Expect(original.Priority()).To(Equal(0))
			Expect(prioritized.Priority()).To(Equal(5))
			Expect(composite.Priority()).To(Equal(0))
			Expect(prioritizedComposite.Priority()).To(Equal(7))
		})
	})

	When("a child transformer returns an error", func() {
		var visitErr error

		BeforeEach(func() {
			matcher, err := path.CompilePath("child")
			Expect(err).NotTo(HaveOccurred())
			transform := NewTransformer(matcher, func(any) (any, error) { return nil, fmt.Errorf("failed") })
			_, visitErr = Visit(loadYAML("child_error/input.yaml"), transform)
		})

		It("reports the failing path", func() { Expect(visitErr).To(MatchError(ContainSubstring("child"))) })
	})

	When("a transform removes a value before later transforms run", func() {
		var called bool

		BeforeEach(func() {
			called = false
			_, err := Visit("value",
				NewTransformer(nil, func(any) (any, error) { return RemoveValue, nil }),
				NewTransformer(nil, func(value any) (any, error) { called = true; return value, nil }),
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not call later transforms", func() { Expect(called).To(BeFalse()) })
	})
})
