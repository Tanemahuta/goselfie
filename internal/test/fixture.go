// Package test provides reusable test fixtures for goselfie packages.
package test

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/onsi/ginkgo/v2"
)

const testDataDir = "__testdata__"

// PrepareTestDir copies one named directory from the calling package's
// __testdata__ directory into a fresh temporary directory.
func PrepareTestDir(name string) string {
	ginkgo.GinkgoHelper()
	root, err := callerTestDataDir()
	if err != nil {
		panic(err)
	}
	if name == "" || filepath.Base(name) != name {
		panic(fmt.Errorf("prepare test directory: invalid fixture name %q", name))
	}
	source := filepath.Join(root, name)
	if info, err := os.Stat(source); err != nil || !info.IsDir() {
		panic(fmt.Errorf("prepare test directory %q: fixture does not exist", name))
	}
	target := ginkgo.GinkgoT().TempDir()
	if err := copyDir(source, target); err != nil {
		panic(fmt.Errorf("prepare test directory: %w", err))
	}
	return target
}

func callerTestDataDir() (string, error) {
	pcs := make([]uintptr, 32)
	count := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:count])
	for {
		frame, more := frames.Next()
		if !strings.HasSuffix(filepath.ToSlash(frame.File), "/internal/test/fixture.go") {
			path := filepath.Join(filepath.Dir(frame.File), testDataDir)
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				return path, nil
			}
		}
		if !more {
			break
		}
	}
	return "", fmt.Errorf("resolve %s for test caller", testDataDir)
}

func copyDir(source, target string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, relative)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		return copyFile(path, destination, info.Mode())
	})
}

func copyFile(source, destination string, mode fs.FileMode) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		_ = input.Close()
		return err
	}
	_, copyErr := io.Copy(output, input)
	inputErr := input.Close()
	outputErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	if inputErr != nil {
		return inputErr
	}
	return outputErr
}
