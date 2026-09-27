package ginkgo

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sync"

	"github.com/tanemahuta/goselfie/snapshot"
	"github.com/tanemahuta/goselfie/utils"
)

type parsedSource struct {
	files *token.FileSet
	file  *ast.File
}

type detectedUpdateModes struct {
	file       snapshot.UpdateMode
	containers map[int]snapshot.UpdateMode
}

//nolint:gochecknoglobals // Parsed directives are cached by test file for the process lifetime.
var updateModeCache = struct {
	sync.Mutex
	byPath map[string]detectedUpdateModes
}{byPath: make(map[string]detectedUpdateModes)}

func parseSource(path string) (*parsedSource, error) {
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse update mode in %s: %w", path, err)
	}
	return &parsedSource{files: files, file: file}, nil
}

func detectFileUpdateMode(path string) (snapshot.UpdateMode, error) {
	modes, err := detectUpdateModes(path)
	if err != nil {
		return snapshot.UpdateModeNever, err
	}
	return modes.file, nil
}

func detectContainerUpdateMode(path string, line int) (snapshot.UpdateMode, error) {
	modes, err := detectUpdateModes(path)
	if err != nil {
		return snapshot.UpdateModeNever, err
	}
	mode, ok := modes.containers[line]
	if !ok {
		return snapshot.UpdateModeMissing, fmt.Errorf("detect container update mode: no Ginkgo container at line %d", line)
	}
	return mode, nil
}

func detectUpdateModes(path string) (detectedUpdateModes, error) {
	updateModeCache.Lock()
	defer updateModeCache.Unlock()
	if cached, ok := updateModeCache.byPath[path]; ok {
		return cached, nil
	}
	source, err := parseSource(path)
	if err != nil {
		return detectedUpdateModes{}, err
	}
	selectedByLine := make(map[int]*ast.CallExpr)
	ast.Inspect(source.file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || !ginkgoContainer(call.Fun) {
			return true
		}
		line := source.files.Position(call.Pos()).Line
		if selected := selectedByLine[line]; selected == nil || call.End() > selected.End() {
			selectedByLine[line] = call
		}
		return true
	})
	modes := detectedUpdateModes{
		file:       source.mode(token.Pos(1), source.file.Package),
		containers: make(map[int]snapshot.UpdateMode, len(selectedByLine)),
	}
	for line, selected := range selectedByLine {
		modes.containers[line] = source.containerMode(selected)
	}
	updateModeCache.byPath[path] = modes
	return modes, nil
}

func (source *parsedSource) containerMode(container *ast.CallExpr) snapshot.UpdateMode {
	excluded := source.matcherExpressions(container)
	ast.Inspect(container, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || call == container || !ginkgoContainer(call.Fun) {
			return true
		}
		excluded = append(excluded, utils.Range[int]{
			Start:   source.offset(call.Pos()),
			EndExcl: source.offset(call.End()),
		})
		return true
	})
	return source.modeExcluding(container.Pos(), container.End(), excluded)
}

// UpdateModeForSource detects directives in the matcher expression containing sourceRange.
func UpdateModeForSource(path string, sourceRange utils.Range[int]) (snapshot.UpdateMode, error) {
	source, err := parseSource(path)
	if err != nil {
		return snapshot.UpdateModeNever, err
	}
	selected := source.matcherExpression(sourceRange)
	if selected == nil {
		return snapshot.UpdateModeMissing, nil
	}
	return source.mode(selected.Pos(), selected.End()), nil
}

func (source *parsedSource) matcherExpression(sourceRange utils.Range[int]) *ast.CallExpr {
	var selected *ast.CallExpr
	ast.Inspect(source.file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || ginkgoContainer(call.Fun) {
			return true
		}
		start := source.offset(call.Pos())
		end := source.offset(call.End())
		if start <= sourceRange.Start && sourceRange.EndExcl <= end && (selected == nil || call.Pos() < selected.Pos()) {
			selected = call
		}
		return true
	})
	return selected
}

func (source *parsedSource) matcherExpressions(container *ast.CallExpr) []utils.Range[int] {
	var expressions []utils.Range[int]
	ast.Inspect(container, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || !snapshotMatcher(call.Fun) {
			return true
		}
		rangeOfCall := utils.Range[int]{Start: source.offset(call.Pos()), EndExcl: source.offset(call.End())}
		if expression := source.matcherExpression(rangeOfCall); expression != nil {
			expressions = append(expressions, utils.Range[int]{Start: source.offset(expression.Pos()), EndExcl: source.offset(expression.End())})
		}
		return true
	})
	return expressions
}

func (source *parsedSource) mode(start, end token.Pos) snapshot.UpdateMode {
	return source.modeExcluding(start, end, nil)
}

func (source *parsedSource) modeExcluding(start, end token.Pos, excluded []utils.Range[int]) snapshot.UpdateMode {
	mode := snapshot.UpdateModeMissing
	for _, group := range source.file.Comments {
		if group.Pos() < start || group.End() > end {
			continue
		}
		commentStart := source.offset(group.Pos())
		commentEnd := source.offset(group.End())
		if containedBy(commentStart, commentEnd, excluded) {
			continue
		}
		text := group.Text()
		if selfieWriteDirective.MatchString(text) {
			return snapshot.UpdateModeAlways
		}
		if selfieOnceDirective.MatchString(text) {
			mode = snapshot.UpdateModeOnce
		}
	}
	return mode
}

func containedBy(start, end int, ranges []utils.Range[int]) bool {
	for _, current := range ranges {
		if current.Start <= start && end <= current.EndExcl {
			return true
		}
	}
	return false
}

func (source *parsedSource) offset(position token.Pos) int {
	return source.files.File(position).Offset(position)
}

func ginkgoContainer(function ast.Expr) bool {
	name := functionName(function)
	switch name {
	case "Describe", "Context", "When", "It", "Specify", "DescribeTable", "Entry":
		return true
	default:
		return false
	}
}

func snapshotMatcher(function ast.Expr) bool {
	name := functionName(function)
	return name == "MatchSnapshot" || name == "MatchSnapshot_TODO"
}

func functionName(function ast.Expr) string {
	switch value := function.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.SelectorExpr:
		return value.Sel.Name
	default:
		return ""
	}
}
