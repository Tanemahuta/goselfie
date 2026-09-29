package gomega

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tanemahuta/goselfie/album"
	selfieginkgo "github.com/tanemahuta/goselfie/ginkgo"
	"github.com/tanemahuta/goselfie/lens"
	"github.com/tanemahuta/goselfie/snapshot"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
	"github.com/tanemahuta/goselfie/utils"
)

type testLens[I any] func(I) ([]byte, error)

func (projection testLens[I]) Apply(value I, _ content.Type) ([]byte, content.Type, error) {
	data, err := projection(value)
	return data, content.Text, err
}

var _ = Describe("Snapshot matcher", func() {
	When("a matcher uses the public default builder", func() {
		var current *matcherTestContext
		var matched bool

		BeforeEach(func() {
			current = matcherTestContextForSource(snapshot.UpdateModeMissing)
			useMatcherTestContext(current)
			projection := testLens[string](func(value string) ([]byte, error) { return []byte(value), nil })
			matcher := MatchSnapshot(NewMatcherBuilder(projection)).WithName("response")
			var err error
			matched, err = matcher.Match("actual")
			Expect(err).NotTo(HaveOccurred())
		})

		It("resolves the current spec and stores the snapshot in its album", func() {
			Expect(matched).To(BeTrue())
			updated := current.snapshotAlbum.UpdatedSnapshots()
			Expect(updated).To(HaveLen(1))
			contents, err := updated[0].Contents()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("actual"))
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
			current.snapshotAlbum.stored[strings.Join(name, "\x00")] = newMatcherSnapshot(name, utils.Range[int]{}, "expected")
			projection := testLens[string](func(value string) ([]byte, error) { return []byte(value), nil })
			matcher := MatchSnapshot(NewUpdatingMatcherBuilder(projection)).WithName("response")
			var err error
			matched, err = matcher.Match("actual")
			Expect(err).NotTo(HaveOccurred())
		})

		It("updates the snapshot without requiring an update directive", func() {
			Expect(matched).To(BeTrue())
			updated := current.snapshotAlbum.UpdatedSnapshots()
			Expect(updated).To(HaveLen(1))
			contents, err := updated[0].Contents()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("actual"))
		})
	})

	When("the current test context cannot be resolved", func() {
		var operation func()

		BeforeEach(func() {
			previous := getTestContext
			getTestContext = func() (selfieginkgo.TestContext, error) { return nil, errors.New("no running spec") }
			DeferCleanup(func() { getTestContext = previous })
			operation = func() {
				NewMatcherBuilder(testLens[string](func(value string) ([]byte, error) { return []byte(value), nil }))
			}
		})

		It("panics with the context error", func() {
			Expect(operation).To(PanicWith(MatchError(ContainSubstring("get current Ginkgo test context"))))
		})
	})

	When("a builder is configured with a name", func() {
		var base, configured MatcherBuilder[string]

		BeforeEach(func() {
			base = NewMatcherBuilderWithContext(newMatcherTestContext(snapshot.UpdateModeMissing), testLens[string](func(value string) ([]byte, error) { return []byte(value), nil }))
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
		stored, actual                string
		expectedMatch, expectedUpdate bool
	}{
		{"missing and disabled", snapshot.UpdateModeNever, snapshot.UpdateModeMissing, "", "new", false, false},
		{"missing and allowed", snapshot.UpdateModeMissing, snapshot.UpdateModeMissing, "", "new", true, true},
		{"matching", snapshot.UpdateModeMissing, snapshot.UpdateModeMissing, "same", "same", true, true},
		{"different and always", snapshot.UpdateModeMissing, snapshot.UpdateModeAlways, "old", "new", true, true},
		{"different and once", snapshot.UpdateModeMissing, snapshot.UpdateModeOnce, "old", "new", true, true},
		{"different without always", snapshot.UpdateModeMissing, snapshot.UpdateModeMissing, "old", "new", false, false},
	} {
		scenario := scenario
		When("update rules apply to "+scenario.name, func() {
			var current *matcherTestContext
			var matched bool

			BeforeEach(func() {
				current = newMatcherTestContext(scenario.contextMode)
				if scenario.stored != "" {
					current.snapshotAlbum.stored["value"] = newMatcherSnapshot([]string{"value"}, utils.Range[int]{}, scenario.stored)
				}
				matcher := namedMatcher(current, scenario.matcherMode, "value")
				var err error
				matched, err = matcher.Match(scenario.actual)
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
			stored = newMatcherSnapshot([]string{"value"}, utils.Range[int]{}, "actual")
			current.snapshotAlbum.stored["value"] = stored
			matcher = namedMatcher(current, snapshot.UpdateModeMissing, "value")
			matcher.source = utils.Range[int]{Start: 12, EndExcl: 34}
			Expect(matcher.Match("actual")).To(BeTrue())
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
			current.snapshotAlbum.stored["suite\x00sibling It\x00#1"] = newMatcherSnapshot([]string{"suite", "sibling It", "#1"}, utils.Range[int]{}, "sibling")
			matcher = namedMatcher(current, snapshot.UpdateModeMissing, "#1")
			matcher.name = []string{"suite", "current It", "#1"}
			var err error
			matched, err = matcher.Match("current")
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not use the sibling snapshot", func() {
			Expect(matched).To(BeFalse())
			Expect(matcher.FailureMessage(nil)).To(Equal("snapshot does not exist and updates are disabled"))
			Expect(current.snapshotAlbum.updated).To(BeEmpty())
		})
	})

	When("a YAML snapshot differs from the actual value", func() {
		var matcher *snapshotMatcher[any]
		var matched bool

		BeforeEach(func() {
			current := newMatcherTestContext(snapshot.UpdateModeNever)
			current.snapshotAlbum.stored["value"] = newMatcherSnapshotWithContentType(
				[]string{"value"}, content.YAML, "name: expected\n",
			)
			matcher = &snapshotMatcher[any]{
				context: current, projection: lens.ToYAML(),
				updateMode: snapshot.UpdateModeMissing, name: []string{"value"},
			}
			var err error
			matched, err = matcher.Match(map[string]any{"name": "actual"})
			Expect(err).NotTo(HaveOccurred())
		})

		It("uses the YAML-specific mismatch messages", func() {
			Expect(matched).To(BeFalse())
			Expect(matcher.FailureMessage(nil)).To(ContainSubstring("to match YAML snapshot"))
			Expect(matcher.FailureMessage(nil)).To(ContainSubstring("name: expected"))
			Expect(matcher.FailureMessage(nil)).To(ContainSubstring("name: actual"))
			Expect(matcher.NegatedFailureMessage(nil)).To(ContainSubstring("not to match YAML snapshot"))
		})
	})

	When("a projection returns an error", func() {
		var matchErr error

		BeforeEach(func() {
			matcher := NewMatcherBuilderWithContext(
				newMatcherTestContext(snapshot.UpdateModeMissing),
				testLens[string](func(string) ([]byte, error) { return nil, errors.New("conversion failed") }),
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
			matched, err = matcher.Match(map[string]any{"name": "example"})
			Expect(err).NotTo(HaveOccurred())
		})

		It("records the lens content type with the updated snapshot", func() {
			Expect(matched).To(BeTrue())
			actual := matcher.(*snapshotMatcher[any]).context.Album().SnapshotByName([]string{"yaml"})
			Expect(actual.ContentType()).To(Equal(content.YAML))
		})
	})
})

func namedMatcher(current *matcherTestContext, mode snapshot.UpdateMode, name string) *snapshotMatcher[string] {
	projection := testLens[string](func(value string) ([]byte, error) {
		return []byte(value), nil
	})
	return &snapshotMatcher[string]{context: current, projection: projection, contentType: content.Text, updateMode: mode, name: []string{name}}
}

type matcherTestContext struct {
	mode          snapshot.UpdateMode
	testFilePath  string
	snapshotAlbum *matcherTestAlbum
}

func newMatcherTestContext(mode snapshot.UpdateMode) *matcherTestContext {
	current := &matcherTestContext{mode: mode, snapshotAlbum: &matcherTestAlbum{stored: map[string]snapshot.Data{}, updated: map[string]snapshot.Taken{}}}
	DeferCleanup(current.Close)
	return current
}

func (current *matcherTestContext) Close() error { return current.snapshotAlbum.Close() }
func (current *matcherTestContext) TestFilePath() string {
	if current.testFilePath != "" {
		return current.testFilePath
	}
	return "test.go"
}
func (current *matcherTestContext) UpdateMode() snapshot.UpdateMode { return current.mode }
func (current *matcherTestContext) Album() album.Album              { return current.snapshotAlbum }

type matcherTestAlbum struct {
	stored  map[string]snapshot.Data
	updated map[string]snapshot.Taken
}

func (current *matcherTestAlbum) TestFilePath() string { return "test.go" }
func (current *matcherTestAlbum) FilePath() string     { return "test.ss" }
func (current *matcherTestAlbum) SnapshotByName(name []string) snapshot.Data {
	key := strings.Join(name, "\x00")
	if updated := current.updated[key]; updated != nil {
		return updated
	}
	return current.stored[key]
}
func (current *matcherTestAlbum) UpdateSnapshot(value snapshot.Taken) error {
	key := strings.Join(value.Name(), "\x00")
	if current.updated[key] != nil {
		return fmt.Errorf("duplicate snapshot")
	}
	delete(current.stored, key)
	current.updated[key] = value
	return nil
}
func (current *matcherTestAlbum) UpdatedSnapshots() []snapshot.Taken {
	values := make([]snapshot.Taken, 0, len(current.updated))
	for _, value := range current.updated {
		values = append(values, value)
	}
	return values
}
func (current *matcherTestAlbum) AllSnapshots() []snapshot.Data {
	values := make([]snapshot.Data, 0, len(current.stored)+len(current.updated))
	for _, value := range current.stored {
		values = append(values, value)
	}
	for _, value := range current.updated {
		values = append(values, value)
	}
	return values
}
func (current *matcherTestAlbum) Write() error { return nil }
func (current *matcherTestAlbum) Close() error {
	for key, value := range current.stored {
		_ = value.Evict()
		delete(current.stored, key)
	}
	for key, value := range current.updated {
		_ = value.Evict()
		delete(current.updated, key)
	}
	return nil
}

func newMatcherSnapshot(name []string, source utils.Range[int], contents string) snapshot.Taken {
	return newMatcherSnapshotData(name, content.Text, source, contents)
}

func newMatcherSnapshotWithContentType(name []string, contentType content.Type, contents string) snapshot.Taken {
	return newMatcherSnapshotData(name, contentType, utils.Range[int]{}, contents)
}

func newMatcherSnapshotData(name []string, contentType content.Type, source utils.Range[int], contents string) snapshot.Taken {
	value, err := snapshot.NewTaken(name, contentType, source, func(writer io.Writer) error {
		_, err := io.WriteString(writer, contents)
		return err
	})
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(value.Evict)
	return value
}

// MatchSnapshot keeps the call shape expected by matcher source detection in these constructor tests.
func MatchSnapshot[I any](builder MatcherBuilder[I]) MatcherBuilder[I] { return builder }

func matcherTestContextForSource(mode snapshot.UpdateMode) *matcherTestContext {
	_, sourcePath, _, ok := runtime.Caller(0)
	Expect(ok).To(BeTrue())
	absolutePath, err := filepath.Abs(sourcePath)
	Expect(err).NotTo(HaveOccurred())
	current := newMatcherTestContext(mode)
	current.testFilePath = absolutePath
	return current
}

func useMatcherTestContext(current *matcherTestContext) {
	previous := getTestContext
	getTestContext = func() (selfieginkgo.TestContext, error) { return current, nil }
	DeferCleanup(func() { getTestContext = previous })
}
