package ginkgo

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	onsiginkgo "github.com/onsi/ginkgo/v2"

	"github.com/tanemahuta/goselfie/album"
	"github.com/tanemahuta/goselfie/snapshot"
)

type testContextCache struct {
	cache map[string]TestContext
	sync.Mutex
}

func newTestContextCache() *testContextCache {
	return &testContextCache{
		cache: make(map[string]TestContext),
	}
}

func (t *testContextCache) Get(path string) (TestContext, error) {
	t.Lock()
	defer t.Unlock()
	result, ok := t.cache[path]
	if !ok {
		var err error
		result, err = openTestContext(path)
		if err != nil {
			return nil, err
		}
		t.cache[path] = result
	}
	return result, nil
}

func (t *testContextCache) closeAll() error {
	t.Lock()
	defer t.Unlock()
	var errs []error
	for path, current := range t.cache {
		if err := current.Close(); err != nil {
			errs = append(errs, err)
		}
		delete(t.cache, path)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

var testContexts = newTestContextCache() //nolint:gochecknoglobals // Ginkgo specs share per-file contexts within the process.

// CurrentTestContext returns the cached context for the current test file with
// the update mode of the running spec's container hierarchy applied.
func CurrentTestContext() (TestContext, error) {
	path, err := detectCurrentTestFile()
	if err != nil {
		return nil, err
	}
	current, err := testContexts.Get(path)
	if err != nil {
		return nil, err
	}
	mode := current.UpdateMode()
	for _, line := range detectCurrentContainerLines() {
		if line <= 0 {
			continue
		}
		containerMode, err := detectContainerUpdateMode(path, line)
		if err != nil {
			return nil, fmt.Errorf("detect update mode for Ginkgo container at line %d: %w", line, err)
		}
		mode = snapshot.CoerceMode(mode, containerMode)
	}
	return specTestContext{TestContext: current, updateMode: mode}, nil
}

type specTestContext struct {
	TestContext
	updateMode snapshot.UpdateMode
}

func (current specTestContext) UpdateMode() snapshot.UpdateMode { return current.updateMode }

// Detect the current test file
func detectCurrentTestFile() (string, error) {
	path := onsiginkgo.CurrentSpecReport().FileName()
	if path == "" {
		return "", errors.New("detect Ginkgo test file: no spec is running")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("detect Ginkgo test file: %w", err)
	}
	return absolute, nil
}

func detectCurrentContainerLines() []int {
	report := onsiginkgo.CurrentSpecReport()
	lines := make([]int, 0, len(report.ContainerHierarchyLocations)+1)
	for _, location := range report.ContainerHierarchyLocations {
		lines = append(lines, location.LineNumber)
	}
	return append(lines, report.LeafNodeLocation.LineNumber)
}

func openTestContext(testFilePath string) (TestContext, error) {
	absolute, err := filepath.Abs(testFilePath)
	if err != nil {
		return nil, fmt.Errorf("resolve ginkgo test file: %w", err)
	}
	mode, err := detectFileUpdateMode(absolute)
	if err != nil {
		return nil, err
	}
	snapshotAlbum, err := album.OpenAlbum(absolute)
	if err != nil {
		return nil, fmt.Errorf("open snapshot album: %w", err)
	}
	return &testContext{
		testFilePath: absolute,
		updateMode:   snapshot.CoerceMode(UpdateMode(), mode),
		album:        snapshotAlbum,
		sourceEditor: NewSourceEditor(absolute),
	}, nil
}

//nolint:gochecknoinits // Register the suite-wide cleanup hook when this package is imported.
func init() {
	onsiginkgo.ReportAfterSuite("close cached ginkgo test contexts", func(onsiginkgo.Report) {
		if err := testContexts.closeAll(); err != nil {
			onsiginkgo.Fail(fmt.Sprintf("close cached Ginkgo test contexts: %v", err))
		}
	})
}
