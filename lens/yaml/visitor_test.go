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

		It("visits children before their parent and keeps the replacement", func() {
			Expect(result).To(Equal(loadYAML("replace_parent/expected.yaml")))
			Expect(calls).To(Equal([]string{"replace.child", "replace", "<root>"}))
		})
	})

	When("order is ignored for a list", func() {
		var result any

		BeforeEach(func() {
			result = visitYAML("ignore_order/list/input.yaml",
				IgnoreOrder("items"), IgnoreOrder("items.3"), IgnoreOrder("plain"), Mask("items.*.token"))
		})

		It("sorts the matched lists by YAML after transforming children and leaves the rest alone", func() {
			Expect(result).To(Equal(loadYAML("ignore_order/list/expected.yaml")))
		})
	})

	When("order is ignored for nested lists", func() {
		const input = "ignore_order/nested_lists/input.yaml"

		It("sorts the outer list by its elements' YAML without touching inner lists", func() {
			Expect(visitYAML(input, IgnoreOrder("matrix"))).
				To(Equal(loadYAML("ignore_order/nested_lists/expected_outer.yaml")))
		})

		It("sorts inner lists before the outer list", func() {
			Expect(visitYAML(input, IgnoreOrder("matrix"), IgnoreOrder("matrix.*"))).
				To(Equal(loadYAML("ignore_order/nested_lists/expected_nested.yaml")))
		})

		It("sorts every list for a recursive path", func() {
			Expect(visitYAML(input, IgnoreOrder("**"))).
				To(Equal(loadYAML("ignore_order/nested_lists/expected_recursive.yaml")))
		})
	})

	When("nested elements differ only in inner order", func() {
		var first, second any

		BeforeEach(func() {
			transforms := []Transformer{IgnoreOrder("items"), IgnoreOrder("items.*.ports")}
			first = visitYAML("ignore_order/inner_order/first.yaml", transforms...)
			second = visitYAML("ignore_order/inner_order/second.yaml", transforms...)
		})

		It("sorts the outer list by the normalized elements", func() {
			expected := loadYAML("ignore_order/inner_order/expected.yaml")
			Expect(first).To(Equal(expected))
			Expect(second).To(Equal(expected))
		})
	})

	When("a container is empty without removal", func() {
		var result any

		BeforeEach(func() { result = visitYAML("empty_containers/input.yaml", Mask("value")) })

		It("keeps the empty containers", func() {
			Expect(result).To(Equal(loadYAML("empty_containers/expected.yaml")))
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
