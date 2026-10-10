package gomega

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tanemahuta/goselfie/album"
	selfieginkgo "github.com/tanemahuta/goselfie/ginkgo"
	testfixture "github.com/tanemahuta/goselfie/internal/test"
	"github.com/tanemahuta/goselfie/snapshot"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
	"github.com/tanemahuta/goselfie/utils"
	sigsyaml "sigs.k8s.io/yaml"
)

type testLens[I any] func(I) ([]byte, error)

func (projection testLens[I]) Apply(value I, _ content.Type) ([]byte, content.Type, error) {
	data, err := projection(value)
	return data, content.Text, err
}

func namedMatcher(current *matcherTestContext, mode snapshot.UpdateMode, name string) *snapshotMatcher[string] {
	projection := matcherTextLens()
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
	_, sourcePath, _, ok := runtime.Caller(1)
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

func matcherYAMLFixture(name string) any {
	var value any
	Expect(sigsyaml.Unmarshal(testfixture.ReadTestFile(name), &value)).To(Succeed())
	return value
}

func matcherTextLens() testLens[string] {
	return func(value string) ([]byte, error) { return []byte(value), nil }
}

func matcherFailingLens(err error) testLens[string] {
	return func(string) ([]byte, error) { return nil, err }
}
