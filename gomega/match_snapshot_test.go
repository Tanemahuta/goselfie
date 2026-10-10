package gomega

import (
	"errors"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	selfieginkgo "github.com/tanemahuta/goselfie/ginkgo"
	testfixture "github.com/tanemahuta/goselfie/internal/test"
	"github.com/tanemahuta/goselfie/lens"
	"github.com/tanemahuta/goselfie/snapshot"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
	"github.com/tanemahuta/goselfie/utils"
)

var _ = Describe("Snapshot matcher", func() {
	When("a matcher uses the public default builder", func() {
		var current *matcherTestContext
		var matched bool

		BeforeEach(func() {
			current = matcherTestContextForSource(snapshot.UpdateModeMissing)
			useMatcherTestContext(current)
			projection := matcherTextLens()
			matcher := MatchSnapshot(NewMatcherBuilder(projection)).WithName("response")
			var err error
			matched, err = matcher.Match(string(testfixture.ReadTestFile("matcher/actual.txt")))
			Expect(err).NotTo(HaveOccurred())
		})

		It("resolves the current spec and stores the snapshot in its album", func() {
			Expect(matched).To(BeTrue())
			updated := current.snapshotAlbum.UpdatedSnapshots()
			Expect(updated).To(HaveLen(1))
			contents, err := updated[0].Contents()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal(string(testfixture.ReadTestFile("matcher/actual.txt"))))
		})
	})

	When("the updating builder sees a changed snapshot", func() {
		var current *matcherTestContext
		var matched bool

		BeforeEach(func() {
			current = matcherTestContextForSource(snapshot.UpdateModeMissing)
			useMatcherTestContext(current)
			report := CurrentSpecReport()
			name := append([]string(nil), report.ContainerHierarchyTexts...)
			name = append(name, report.LeafNodeText, "response")
			current.snapshotAlbum.stored[strings.Join(name, "\x00")] = newMatcherSnapshot(name, utils.Range[int]{}, string(testfixture.ReadTestFile("matcher/expected.txt")))
			projection := matcherTextLens()
			matcher := MatchSnapshot(NewUpdatingMatcherBuilder(projection)).WithName("response")
			var err error
			matched, err = matcher.Match(string(testfixture.ReadTestFile("matcher/actual.txt")))
			Expect(err).NotTo(HaveOccurred())
		})

		It("updates the snapshot without requiring an update directive", func() {
			Expect(matched).To(BeTrue())
			updated := current.snapshotAlbum.UpdatedSnapshots()
			Expect(updated).To(HaveLen(1))
			contents, err := updated[0].Contents()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal(string(testfixture.ReadTestFile("matcher/actual.txt"))))
		})
	})

	When("the current test context cannot be resolved", func() {
		var operation func()

		BeforeEach(func() {
			previous := getTestContext
			getTestContext = func() (selfieginkgo.TestContext, error) { return nil, errors.New("no running spec") }
			DeferCleanup(func() { getTestContext = previous })
			operation = func() {
				NewMatcherBuilder(matcherTextLens())
			}
		})

		It("panics with the context error", func() {
			Expect(operation).To(PanicWith(MatchError(ContainSubstring("get current Ginkgo test context"))))
		})
	})

	When("a builder is configured with a name", func() {
		var base, configured MatcherBuilder[string]

		BeforeEach(func() {
			base = NewMatcherBuilderWithContext(newMatcherTestContext(snapshot.UpdateModeMissing), matcherTextLens())
			configured = base.WithName("response")
		})

		It("returns a clone without changing the original", func() {
			Expect(configured).NotTo(BeIdenticalTo(base))
			Expect(base.(*snapshotMatcher[string]).name).To(Equal([]string{"#1"}))
			Expect(configured.(*snapshotMatcher[string]).name).To(Equal([]string{"response"}))
		})
	})

	for _, scenario := range []struct {
		name                          string
		contextMode, matcherMode      snapshot.UpdateMode
		storedFixture, actualFixture  string
		expectedMatch, expectedUpdate bool
	}{
		{"missing and disabled", snapshot.UpdateModeNever, snapshot.UpdateModeMissing, "", "update_rules/new.txt", false, false},
		{"missing and allowed", snapshot.UpdateModeMissing, snapshot.UpdateModeMissing, "", "update_rules/new.txt", true, true},
		{"matching", snapshot.UpdateModeMissing, snapshot.UpdateModeMissing, "update_rules/same.txt", "update_rules/same.txt", true, true},
		{"different and always", snapshot.UpdateModeMissing, snapshot.UpdateModeAlways, "update_rules/old.txt", "update_rules/new.txt", true, true},
		{"different and once", snapshot.UpdateModeMissing, snapshot.UpdateModeOnce, "update_rules/old.txt", "update_rules/new.txt", true, true},
		{"different without always", snapshot.UpdateModeMissing, snapshot.UpdateModeMissing, "update_rules/old.txt", "update_rules/new.txt", false, false},
	} {
		scenario := scenario
		When("update rules apply to "+scenario.name, func() {
			var current *matcherTestContext
			var matched bool

			BeforeEach(func() {
				current = newMatcherTestContext(scenario.contextMode)
				if scenario.storedFixture != "" {
					current.snapshotAlbum.stored["value"] = newMatcherSnapshot([]string{"value"}, utils.Range[int]{}, string(testfixture.ReadTestFile(scenario.storedFixture)))
				}
				matcher := namedMatcher(current, scenario.matcherMode, "value")
				var err error
				matched, err = matcher.Match(string(testfixture.ReadTestFile(scenario.actualFixture)))
				Expect(err).NotTo(HaveOccurred())
			})

			It("matches and updates according to the selected modes", func() {
				Expect(matched).To(Equal(scenario.expectedMatch))
				Expect(current.snapshotAlbum.updated["value"] != nil).To(Equal(scenario.expectedUpdate))
			})
		})
	}

	When("the actual value matches a stored snapshot", func() {
		var current *matcherTestContext
		var stored snapshot.Taken
		var matcher *snapshotMatcher[string]

		BeforeEach(func() {
			current = newMatcherTestContext(snapshot.UpdateModeMissing)
			stored = newMatcherSnapshot([]string{"value"}, utils.Range[int]{}, string(testfixture.ReadTestFile("matcher/actual.txt")))
			current.snapshotAlbum.stored["value"] = stored
			matcher = namedMatcher(current, snapshot.UpdateModeMissing, "value")
			matcher.source = utils.Range[int]{Start: 12, EndExcl: 34}
			Expect(matcher.Match(string(testfixture.ReadTestFile("matcher/actual.txt")))).To(BeTrue())
		})

		It("retains the stored snapshot with its matcher source", func() {
			Expect(current.snapshotAlbum.updated["value"]).To(BeIdenticalTo(stored))
			Expect(current.snapshotAlbum.updated["value"].Source()).To(Equal(matcher.source))
		})
	})

	When("a sibling It has the same expectation number", func() {
		var current *matcherTestContext
		var matcher *snapshotMatcher[string]
		var matched bool

		BeforeEach(func() {
			current = newMatcherTestContext(snapshot.UpdateModeNever)
			current.snapshotAlbum.stored["suite\x00sibling It\x00#1"] = newMatcherSnapshot([]string{"suite", "sibling It", "#1"}, utils.Range[int]{}, string(testfixture.ReadTestFile("matcher/sibling.txt")))
			matcher = namedMatcher(current, snapshot.UpdateModeMissing, "#1")
			matcher.name = []string{"suite", "current It", "#1"}
			var err error
			matched, err = matcher.Match(string(testfixture.ReadTestFile("matcher/current.txt")))
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not use the sibling snapshot", func() {
			Expect(matched).To(BeFalse())
			Expect(matcher.FailureMessage(nil)).To(Equal(string(testfixture.ReadTestFile("missing/expected_message.txt"))))
			Expect(current.snapshotAlbum.updated).To(BeEmpty())
		})
	})

	When("a YAML snapshot differs from the actual value", func() {
		var matcher *snapshotMatcher[any]
		var matched bool

		BeforeEach(func() {
			current := newMatcherTestContext(snapshot.UpdateModeNever)
			current.snapshotAlbum.stored["value"] = newMatcherSnapshotWithContentType(
				[]string{"value"}, content.YAML, string(testfixture.ReadTestFile("yaml_mismatch/expected.yaml")),
			)
			matcher = &snapshotMatcher[any]{
				context: current, projection: lens.ToYAML(),
				updateMode: snapshot.UpdateModeMissing, name: []string{"value"},
			}
			var err error
			matched, err = matcher.Match(matcherYAMLFixture("yaml_mismatch/actual.yaml"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("uses the YAML-specific mismatch messages", func() {
			Expect(matched).To(BeFalse())
			Expect(matcher.FailureMessage(nil)).To(ContainSubstring("to match YAML snapshot"))
			Expect(matcher.FailureMessage(nil)).To(ContainSubstring(string(testfixture.ReadTestFile("yaml_mismatch/expected_fragment.txt"))))
			Expect(matcher.FailureMessage(nil)).To(ContainSubstring(string(testfixture.ReadTestFile("yaml_mismatch/actual_fragment.txt"))))
			Expect(matcher.NegatedFailureMessage(nil)).To(ContainSubstring("not to match YAML snapshot"))
		})
	})

	When("a projection returns an error", func() {
		var matchErr error

		BeforeEach(func() {
			matcher := NewMatcherBuilderWithContext(
				newMatcherTestContext(snapshot.UpdateModeMissing),
				matcherFailingLens(errors.New("conversion failed")),
			)
			_, matchErr = matcher.Match("value")
		})

		It("returns the projection error", func() { Expect(matchErr).To(MatchError(ContainSubstring("conversion failed"))) })
	})

	When("a matcher uses a typed YAML lens", func() {
		var matcher MatcherBuilder[any]
		var matched bool

		BeforeEach(func() {
			matcher = NewMatcherBuilderWithContext(newMatcherTestContext(snapshot.UpdateModeMissing), lens.ToYAML()).WithName("yaml")
			var err error
			matched, err = matcher.Match(matcherYAMLFixture("typed_yaml/input.yaml"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("records the lens content type with the updated snapshot", func() {
			Expect(matched).To(BeTrue())
			actual := matcher.(*snapshotMatcher[any]).context.Album().SnapshotByName([]string{"yaml"})
			Expect(actual.ContentType()).To(Equal(content.YAML))
		})
	})
})
