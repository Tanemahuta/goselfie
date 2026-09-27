package ginkgo

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/tanemahuta/goselfie/snapshot"
)

// SourceEditor applies source changes for snapshots updated by a test.
type SourceEditor interface {
	// Cleanup removes one-shot directives for successfully updated snapshots.
	Cleanup(updated []snapshot.Taken) error
}

type sourceEditor struct{ path string }

// NewSourceEditor decorates a Go source file with snapshot cleanup behavior.
func NewSourceEditor(path string) SourceEditor {
	return &sourceEditor{path: path}
}

func (editor *sourceEditor) Cleanup(updated []snapshot.Taken) error {
	return cleanupSource(editor.path, updated)
}

var (
	selfieWriteDirective = regexp.MustCompile(`\bSELFIEWRITE\b`)
	selfieOnceDirective  = regexp.MustCompile(`\bselfieonce\b`)
)

type sourceEdit struct{ start, end int }

func cleanupSource(path string, updated []snapshot.Taken) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read Ginkgo test source: %w", err)
	}
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, path, data, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse Ginkgo test source: %w", err)
	}
	var edits []sourceEdit
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if selfieOnceDirective.MatchString(comment.Text) {
				edits = append(edits, sourceEdit{sourceOffset(files, comment.Pos()), sourceOffset(files, comment.End())})
			}
		}
	}
	ast.Inspect(file, func(node ast.Node) bool {
		identifier, ok := node.(*ast.Ident)
		if !ok || identifier.Name != "MatchSnapshot_TODO" {
			return true
		}
		start := sourceOffset(files, identifier.Pos())
		end := sourceOffset(files, identifier.End())
		if insideUpdatedSnapshot(start, end, updated) {
			edits = append(edits, sourceEdit{end - len("_TODO"), end})
		}
		return true
	})
	if len(edits) == 0 {
		return nil
	}
	sort.Slice(edits, func(left, right int) bool { return edits[left].start > edits[right].start })
	for _, edit := range edits {
		data = append(data[:edit.start], data[edit.end:]...)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat Ginkgo test source: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".goselfie-source-*")
	if err != nil {
		return fmt.Errorf("create temporary Ginkgo test source: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(info.Mode()); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace Ginkgo test source: %w", err)
	}
	return nil
}

func insideUpdatedSnapshot(start, end int, updated []snapshot.Taken) bool {
	for _, current := range updated {
		source := current.Source()
		if source.Start <= start && end <= source.EndExcl {
			return true
		}
	}
	return false
}

func sourceOffset(files *token.FileSet, position token.Pos) int {
	return files.File(position).Offset(position)
}
