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
			transform := NewTransformer(func(path []string) bool {
				name := formatPath(path)
				calls = append(calls, name)
				return name == "replace"
			}, func(any) (any, error) { return map[string]any{"child": "new"}, nil })
			var err error
			result, err = Visit(map[string]any{"replace": map[string]any{"child": "old"}}, transform)
			Expect(err).NotTo(HaveOccurred())
		})

		It("keeps the replacement and skips its child", func() {
			Expect(result).To(Equal(map[string]any{"replace": map[string]any{"child": "new"}}))
			Expect(calls).NotTo(ContainElement("replace.child"))
		})
	})

	When("a transform returns nil or the removal marker", func() {
		var result any

		BeforeEach(func() {
			value := map[string]any{"nullable": "replace", "discard": "remove"}
			transform := NewTransformer(
				func(path []string) bool { return formatPath(path) == "nullable" },
				func(any) (any, error) { return nil, nil },
			)
			var err error
			result, err = Visit(value, transform, Remove("discard"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("preserves nil and removes the marked value", func() {
			Expect(result).To(Equal(map[string]any{"nullable": nil}))
		})
	})

	When("multiple list elements are removed", func() {
		var result any

		BeforeEach(func() {
			value := map[string]any{"items": []any{"first", "keep", "third"}}
			var err error
			result, err = Visit(value, Remove("items.0"), Remove("items.2"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("removes original indexes without skipping shifted values", func() {
			Expect(result).To(Equal(map[string]any{"items": []any{"keep"}}))
		})
	})

	When("a transform removes the root", func() {
		var result any

		BeforeEach(func() {
			transform := NewTransformer(
				func(path []string) bool { return len(path) == 0 },
				func(any) (any, error) { return RemoveValue, nil },
			)
			var err error
			result, err = Visit(map[string]any{"value": 1}, transform)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the removal marker", func() { Expect(IsRemove(result)).To(BeTrue()) })
	})

	When("nil transforms appear in nested combinations", func() {
		var unchanged, changed any

		BeforeEach(func() {
			combined := Combine(nil, Combine(nil, Mask("value"), nil), nil)
			var err error
			unchanged, err = Visit(map[string]any{"empty": map[string]any{}}, nil, Combine(nil))
			Expect(err).NotTo(HaveOccurred())
			changed, err = Visit(map[string]any{"value": "secret"}, combined)
			Expect(err).NotTo(HaveOccurred())
		})

		It("ignores nil transforms", func() {
			Expect(Combine(nil, Combine(nil))).To(BeNil())
			Expect(unchanged).To(HaveKey("empty"))
			Expect(changed).To(Equal(map[string]any{"value": "<masked>"}))
		})
	})

	When("flat and nested transforms have different priorities", func() {
		var flatCalls, nestedCalls []string

		BeforeEach(func() {
			record := func(calls *[]string, name string, priority int) Transformer {
				return NewTransformer(nil, func(value any) (any, error) {
					*calls = append(*calls, name)
					return value, nil
				}).WithPriority(priority)
			}
			_, err := Visit("value",
				record(&flatCalls, "last", 10),
				record(&flatCalls, "first tie", 0),
				record(&flatCalls, "first", -10),
				record(&flatCalls, "second tie", 0),
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = Visit("value", Combine(record(&nestedCalls, "late", 10), record(&nestedCalls, "early", -10)), record(&nestedCalls, "middle", 0))
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

		BeforeEach(func() {
			transform := Remove("spec..uid")
			_, applyErr = transform.Apply([]string{"spec", "uid"}, "value")
		})

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
			_, visitErr = Visit(map[string]any{"child": "value"}, transform)
		})

		It("reports the failing path", func() { Expect(visitErr).To(MatchError(ContainSubstring("child"))) })
	})

	When("a transform removes a value before later transforms run", func() {
		var called bool

		BeforeEach(func() {
			transforms := []Transformer{
				NewTransformer(nil, func(any) (any, error) { return RemoveValue, nil }),
				NewTransformer(nil, func(value any) (any, error) { called = true; return value, nil }),
			}
			_, err := Visit("value", transforms...)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not call later transforms", func() { Expect(called).To(BeFalse()) })
	})
})
