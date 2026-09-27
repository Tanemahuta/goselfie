package path

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Path matcher", func() {
	When("a path expression is empty or malformed", func() {
		expressions := []string{"", ".a", "a..b", "a.", "'unterminated", "a.'unterminated"}
		var results []Matcher
		var errs []error

		BeforeEach(func() {
			results = nil
			errs = nil
			for _, expression := range expressions {
				matcher, err := CompilePath(expression)
				results, errs = append(results, matcher), append(errs, err)
			}
		})

		It("rejects every expression", func() {
			for index := range expressions {
				Expect(errs[index]).To(HaveOccurred())
				Expect(results[index]).To(BeNil())
			}
		})
	})

	When("quoted keys contain dots and escaped quotes", func() {
		var matches []bool

		BeforeEach(func() {
			scenarios := []struct {
				expression string
				path       []string
			}{
				{"labels.'example.com/key'", []string{"labels", "example.com/key"}},
				{"labels.\"a.b\"", []string{"labels", "a.b"}},
				{"labels.'it\\'s'", []string{"labels", "it's"}},
			}
			matches = nil
			for _, scenario := range scenarios {
				matcher, err := CompilePath(scenario.expression)
				Expect(err).NotTo(HaveOccurred())
				matches = append(matches, matcher(scenario.path))
			}
		})

		It("matches each quoted key", func() { Expect(matches).To(ConsistOf(true, true, true)) })
	})

	for _, scenario := range []struct {
		name, expression   string
		matching, rejected [][]string
	}{
		{"a full path", "metadata.uid", [][]string{{"metadata", "uid"}}, [][]string{{"metadata"}, {"metadata", "uid", "extra"}}},
		{"one wildcard segment", "spec.*.uid", [][]string{{"spec", "item", "uid"}}, [][]string{{"spec", "uid"}, {"spec", "item", "nested", "uid"}}},
		{"zero or more wildcard segments", "spec.**.uid", [][]string{{"spec", "uid"}, {"spec", "item", "uid"}, {"spec", "item", "nested", "uid"}}, nil},
	} {
		scenario := scenario
		When("the expression matches "+scenario.name, func() {
			var matching, rejected []bool

			BeforeEach(func() {
				matcher, err := CompilePath(scenario.expression)
				Expect(err).NotTo(HaveOccurred())
				matching = nil
				rejected = nil
				for _, candidate := range scenario.matching {
					matching = append(matching, matcher(candidate))
				}
				for _, candidate := range scenario.rejected {
					rejected = append(rejected, matcher(candidate))
				}
			})

			It("accepts only paths matching the expression", func() {
				for _, value := range matching {
					Expect(value).To(BeTrue())
				}
				for _, value := range rejected {
					Expect(value).To(BeFalse())
				}
			})
		})
	}

	for _, expression := range []string{"spec.'*'.uid", "spec.\"*\".uid", "spec.'**'.uid", "spec.\"**\".uid"} {
		expression := expression
		When("wildcards are quoted in "+expression, func() {
			var matched bool

			BeforeEach(func() {
				matcher, err := CompilePath(expression)
				Expect(err).NotTo(HaveOccurred())
				matched = matcher([]string{"spec", "item", "uid"})
			})

			It("treats the quoted wildcard as a wildcard", func() { Expect(matched).To(BeTrue()) })
		})
	}
})
