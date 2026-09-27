package gomega

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	onsiginkgo "github.com/onsi/ginkgo/v2"
	selfieginkgo "github.com/tanemahuta/goselfie/ginkgo"
	"github.com/tanemahuta/goselfie/gomega/failuremessage"
	"github.com/tanemahuta/goselfie/lens"
	"github.com/tanemahuta/goselfie/snapshot"
	"github.com/tanemahuta/goselfie/snapshot/content"
	"github.com/tanemahuta/goselfie/utils"
)

var getTestContext = selfieginkgo.CurrentTestContext

// NewMatcherBuilder creates a snapshot matcher for the current Ginkgo spec.
func NewMatcherBuilder[I any](projection lens.Lens[I, []byte]) MatcherBuilder[I] {
	return newCurrentMatcherBuilder(projection, snapshot.UpdateModeMissing)
}

// NewUpdatingMatcherBuilder creates a matcher which always updates unless its
// test context disables updates.
func NewUpdatingMatcherBuilder[I any](projection lens.Lens[I, []byte]) MatcherBuilder[I] {
	return newCurrentMatcherBuilder(projection, snapshot.UpdateModeAlways)
}

func newCurrentMatcherBuilder[I any](projection lens.Lens[I, []byte], mode snapshot.UpdateMode) MatcherBuilder[I] {
	current, err := getTestContext()
	if err != nil {
		panic(fmt.Errorf("get current Ginkgo test context: %w", err))
	}
	source, err := detectMatcherSource(current.TestFilePath())
	if err != nil {
		panic(fmt.Errorf("detect snapshot matcher source: %w", err))
	}
	sourceMode, err := selfieginkgo.UpdateModeForSource(current.TestFilePath(), source)
	if err != nil {
		panic(fmt.Errorf("detect snapshot matcher update mode: %w", err))
	}
	return &snapshotMatcher[I]{
		context:    current,
		projection: projection,
		updateMode: snapshot.CoerceMode(mode, sourceMode),
		source:     source,
		name:       nextSnapshotName(current),
	}
}

// NewMatcherBuilderWithContext creates a matcher with an explicitly supplied
// test context. It is useful for isolated tests and custom integrations.
func NewMatcherBuilderWithContext[I any](current selfieginkgo.TestContext, projection lens.Lens[I, []byte]) MatcherBuilder[I] {
	return &snapshotMatcher[I]{
		context:    current,
		projection: projection,
		updateMode: snapshot.UpdateModeMissing,
		name:       []string{"#1"},
	}
}

type snapshotMatcher[I any] struct {
	context     selfieginkgo.TestContext
	projection  lens.Lens[I, []byte]
	contentType content.Type
	updateMode  snapshot.UpdateMode
	source      utils.Range[int]
	name        []string
	expected    []byte
	actual      []byte
	problem     string
}

func (matcher *snapshotMatcher[I]) WithName(name string) MatcherBuilder[I] {
	clone := matcher.clone()
	if len(clone.name) == 0 {
		clone.name = []string{name}
	} else {
		clone.name[len(clone.name)-1] = name
	}
	return clone
}

func (matcher *snapshotMatcher[I]) clone() *snapshotMatcher[I] {
	return &snapshotMatcher[I]{
		context: matcher.context, projection: matcher.projection, contentType: matcher.contentType,
		updateMode: matcher.updateMode, source: matcher.source,
		name: append([]string(nil), matcher.name...),
	}
}

func (matcher *snapshotMatcher[I]) Match(actual any) (bool, error) {
	matcher.expected = nil
	matcher.actual = nil
	matcher.problem = ""
	if matcher.context == nil {
		return false, fmt.Errorf("ginkgo test context is nil")
	}
	if matcher.projection == nil {
		return false, fmt.Errorf("snapshot projection is nil")
	}
	input, ok := actual.(I)
	if !ok {
		return false, fmt.Errorf("snapshot value of type %T is incompatible with its projection", actual)
	}
	projected, contentType, err := matcher.projection.Apply(input, content.Unknown)
	if err != nil {
		return false, fmt.Errorf("project snapshot: %w", err)
	}
	matcher.actual = projected
	matcher.contentType = contentType
	mode := snapshot.CoerceMode(matcher.context.UpdateMode(), matcher.updateMode)
	snapshotAlbum := matcher.context.Album()
	stored := snapshotAlbum.SnapshotByName(matcher.name)
	if stored == nil {
		if mode == snapshot.UpdateModeNever {
			matcher.problem = "snapshot does not exist and updates are disabled"
			return false, nil
		}
		if err := matcher.update(); err != nil {
			return false, err
		}
		return true, nil
	}
	matcher.expected, err = stored.Contents()
	if err != nil {
		return false, fmt.Errorf("read snapshot: %w", err)
	}
	if bytes.Equal(matcher.expected, matcher.actual) {
		if err := snapshotAlbum.UpdateSnapshot(stored.Promote(matcher.source)); err != nil {
			return false, fmt.Errorf("retain matched snapshot: %w", err)
		}
		return true, nil
	}
	if mode == snapshot.UpdateModeAlways || mode == snapshot.UpdateModeOnce {
		if err := matcher.update(); err != nil {
			return false, err
		}
		return true, nil
	}
	matcher.problem = "snapshot differs"
	return false, nil
}

func (matcher *snapshotMatcher[I]) update() error {
	taken, err := snapshot.NewTaken(matcher.name, matcher.contentType, matcher.source, func(writer io.Writer) error {
		written, err := writer.Write(matcher.actual)
		if err == nil && written != len(matcher.actual) {
			return io.ErrShortWrite
		}
		return err
	})
	if err != nil {
		return fmt.Errorf("take snapshot: %w", err)
	}
	matcher.contentType = taken.ContentType()
	if err := matcher.context.Album().UpdateSnapshot(taken); err != nil {
		_ = taken.Evict()
		return fmt.Errorf("update snapshot: %w", err)
	}
	return nil
}

func (matcher *snapshotMatcher[I]) FailureMessage(_ any) string {
	if matcher.problem == "snapshot differs" {
		return failuremessage.For(matcher.contentType).FailureMessage(matcher.actual, matcher.expected)
	}
	return matcher.problem
}

func (matcher *snapshotMatcher[I]) NegatedFailureMessage(_ any) string {
	return failuremessage.For(matcher.contentType).NegateFailureMessage(matcher.actual, matcher.expected)
}

var snapshotCounters = struct {
	sync.Mutex
	values map[string]int
}{values: make(map[string]int)}

func nextSnapshotName(current selfieginkgo.TestContext) []string {
	report := onsiginkgo.CurrentSpecReport()
	name := append([]string(nil), report.ContainerHierarchyTexts...)
	name = append(name, report.LeafNodeText)
	key := fmt.Sprintf("%p\x00%s\x00%s", current, current.TestFilePath(), strings.Join(name, "\x00"))
	snapshotCounters.Lock()
	snapshotCounters.values[key]++
	counter := snapshotCounters.values[key]
	snapshotCounters.Unlock()
	onsiginkgo.DeferCleanup(func() {
		snapshotCounters.Lock()
		delete(snapshotCounters.values, key)
		snapshotCounters.Unlock()
	})
	return append(name, "#"+strconv.Itoa(counter))
}

func detectMatcherSource(path string) (utils.Range[int], error) {
	line := 0
	pcs := make([]uintptr, 64)
	count := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:count])
	for {
		frame, more := frames.Next()
		if filepath.Clean(frame.File) == filepath.Clean(path) {
			line = frame.Line
			break
		}
		if !more {
			break
		}
	}
	if line == 0 {
		return utils.Range[int]{}, fmt.Errorf("no caller in %s", path)
	}
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, path, nil, 0)
	if err != nil {
		return utils.Range[int]{}, err
	}
	var selected *ast.CallExpr
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || files.Position(call.Pos()).Line > line || files.Position(call.End()).Line < line {
			return true
		}
		if matcherFunction(call.Fun) {
			selected = call
			return false
		}
		return true
	})
	if selected == nil {
		return utils.Range[int]{}, fmt.Errorf("no snapshot matcher call at line %d", line)
	}
	return utils.Range[int]{
		Start:   files.File(selected.Pos()).Offset(selected.Pos()),
		EndExcl: files.File(selected.End()).Offset(selected.End()),
	}, nil
}

func matcherFunction(function ast.Expr) bool {
	var name string
	switch value := function.(type) {
	case *ast.Ident:
		name = value.Name
	case *ast.SelectorExpr:
		name = value.Sel.Name
	}
	return name == "MatchSnapshot" || name == "MatchSnapshot_TODO"
}
