package ginkgo

import (
	"fmt"
	"io"
	"sync"

	"github.com/tanemahuta/goselfie/album"
	"github.com/tanemahuta/goselfie/snapshot"
)

// TestContext owns the snapshot state for one Ginkgo test file.
type TestContext interface {
	io.Closer
	// TestFilePath returns the absolute path of the current test file.
	TestFilePath() string
	// UpdateMode returns the effective update mode for this test file.
	UpdateMode() snapshot.UpdateMode
	// Album returns the snapshots belonging to this test file.
	Album() album.Album
}

type testContext struct {
	testFilePath string
	updateMode   snapshot.UpdateMode
	album        album.Album
	sourceEditor SourceEditor
	closeOnce    sync.Once
	closeErr     error
}

func (current *testContext) TestFilePath() string            { return current.testFilePath }
func (current *testContext) UpdateMode() snapshot.UpdateMode { return current.updateMode }
func (current *testContext) Album() album.Album              { return current.album }

// Close persists the album, cleans update directives, and releases snapshot data.
func (current *testContext) Close() error {
	current.closeOnce.Do(func() {
		defer func() {
			closeErr := current.album.Close()
			if current.closeErr == nil {
				current.closeErr = closeErr
			}
		}()
		if err := current.album.Write(); err != nil {
			current.closeErr = fmt.Errorf("write snapshot album: %w", err)
			return
		}
		if err := current.sourceEditor.Cleanup(current.album.UpdatedSnapshots()); err != nil {
			current.closeErr = err
		}
	})
	return current.closeErr
}
