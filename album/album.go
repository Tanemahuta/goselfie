package album

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/tanemahuta/goselfie/snapshot"
)

// Album represents the collection of snapshots for a test.
type Album interface {
	// TestFilePath returns the file containing the tests that use this album.
	TestFilePath() string
	// FilePath returns the snapshot album's storage path.
	FilePath() string
	// SnapshotByName returns the snapshot with the given name.
	SnapshotByName(name []string) snapshot.Data
	// UpdateSnapshot marks the provided snapshot.Taken as updated.
	UpdateSnapshot(snapshot snapshot.Taken) error
	// UpdatedSnapshots returns snapshots retained or updated in this session.
	UpdatedSnapshots() []snapshot.Taken
	// AllSnapshots returns the snapshots to write, retaining untouched containers
	// and replacing snapshots from containers updated in this session.
	AllSnapshots() []snapshot.Data
	// Write atomically persists all snapshots in the album.
	Write() error
	// Close evicts all temporary snapshot data held by the album.
	Close() error
}

const (
	// SnapshotsDirName is the directory placed next to a test file.
	SnapshotsDirName = "__snapshots__"
	// GoFileExt is the extension removed from a test filename.
	GoFileExt = ".go"
	// SnapshotsFileExt is the extension used by snapshot albums.
	SnapshotsFileExt = ".ss"
)

// OpenAlbum creates a new Album for the provided test file.
func OpenAlbum(testFilePath string) (Album, error) {
	filePath := filepath.Join(
		filepath.Dir(testFilePath),
		SnapshotsDirName,
		strings.TrimSuffix(filepath.Base(testFilePath), GoFileExt)+SnapshotsFileExt,
	)
	result := &album{
		testFilePath: testFilePath,
		filePath:     filePath,
		snapshots:    make(map[string]snapshot.Data),
		updated:      make(map[string]snapshot.Taken),
	}
	if err := result.readSnapshots(); err != nil {
		return nil, err
	}
	return result, nil
}

type album struct {
	testFilePath, filePath string
	snapshots              map[string]snapshot.Data
	updated                map[string]snapshot.Taken
}

func (a *album) FilePath() string {
	return a.filePath
}

func (a *album) TestFilePath() string {
	return a.testFilePath
}

func (a *album) SnapshotByName(name []string) snapshot.Data {
	key := a.nameToKey(name)
	if actual, ok := a.updated[key]; ok {
		return actual
	}
	return a.snapshots[key]
}

func (a *album) UpdateSnapshot(snapshot snapshot.Taken) error {
	if snapshot == nil {
		return errors.New("snapshot is nil")
	}
	key := a.nameToKey(snapshot.Name())
	if previous, ok := a.updated[key]; ok {
		return fmt.Errorf("a snapshot with that name was already taken by %s:%d", a.testFilePath, previous.Source())
	}
	if stored, ok := a.snapshots[key]; ok {
		if !snapshot.Promoted() {
			if err := stored.Evict(); err != nil {
				return fmt.Errorf("evict stored snapshot %q: %w", snapshot.Name(), err)
			}
		}
		delete(a.snapshots, key)
	}
	a.updated[key] = snapshot
	return nil
}

func (a *album) UpdatedSnapshots() []snapshot.Taken {
	result := slices.Collect(maps.Values(a.updated))
	sort.Slice(result, func(left, right int) bool {
		return a.nameToKey(result[left].Name()) < a.nameToKey(result[right].Name())
	})
	return result
}

func (a *album) AllSnapshots() []snapshot.Data {
	updatedContainers := make(map[string]struct{}, len(a.updated))
	for _, updated := range a.updated {
		if container, ok := a.containerKey(updated.Name()); ok {
			updatedContainers[container] = struct{}{}
		}
	}
	all := make(map[string]snapshot.Data, len(a.snapshots)+len(a.updated))
	for key, stored := range a.snapshots {
		if container, ok := a.containerKey(stored.Name()); ok {
			if _, updated := updatedContainers[container]; updated {
				continue
			}
		}
		all[key] = stored
	}
	for key, updated := range a.updated {
		all[key] = updated
	}
	result := slices.Collect(maps.Values(all))
	sort.Slice(result, func(left, right int) bool {
		return a.nameToKey(result[left].Name()) < a.nameToKey(result[right].Name())
	})
	return result
}

func (a *album) containerKey(name []string) (string, bool) {
	if len(name) == 0 {
		return "", false
	}
	return a.nameToKey(name[:len(name)-1]), true
}

func (a *album) Write() error {
	if len(a.updated) == 0 {
		if len(a.snapshots) > 0 {
			return nil
		}
		// No snapshots, no snapshots file.
		if err := os.Remove(a.filePath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	// Create directory and encode to file
	return a.writeSnapshots()
}

func (a *album) Close() error {
	var errs []error
	for key, stored := range a.snapshots {
		if err := stored.Evict(); err != nil {
			errs = append(errs, fmt.Errorf("evict snapshot %s: %w", key, err))
		}
		delete(a.snapshots, key)
	}
	for key, updated := range a.updated {
		if err := updated.Evict(); err != nil {
			errs = append(errs, fmt.Errorf("evict snapshot %s: %w", key, err))
		}
		delete(a.updated, key)
	}
	return errors.Join(errs...)
}

func (a *album) readSnapshots() error {
	file, err := os.Open(a.filePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open snapshots album %s: %w", a.filePath, err)
	}
	decoded, err := Decode(file)
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("decode snapshots album %s: %w", a.filePath, err)
	}
	if err := file.Close(); err != nil {
		evictSnapshots(decoded)
		return fmt.Errorf("close snapshots album %s: %w", a.filePath, err)
	}
	for _, taken := range decoded {
		key := a.nameToKey(taken.Name())
		if _, exists := a.snapshots[key]; exists {
			evictSnapshots(decoded)
			return fmt.Errorf("duplicate snapshot %q in %s", taken.Name(), a.filePath)
		}
		a.snapshots[key] = taken
	}
	return nil
}

func (a *album) nameToKey(name []string) string {
	return strings.Join(name, "\x00")
}

func (a *album) writeSnapshots() error {
	directory := filepath.Dir(a.filePath)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create snapshot directory: %w", err)
	}
	file, err := os.CreateTemp(directory, ".goselfie-album-*")
	if err != nil {
		return fmt.Errorf("create temporary snapshot album: %w", err)
	}
	temporaryPath := file.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err = Encode(a.AllSnapshots(), file); err != nil {
		_ = file.Close()
		return fmt.Errorf("encode snapshot album: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close temporary snapshot album: %w", err)
	}
	if err := os.Rename(temporaryPath, a.filePath); err != nil {
		return fmt.Errorf("replace snapshot album: %w", err)
	}
	return nil
}
