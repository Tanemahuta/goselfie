package ginkgo

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/gomega"

	testfixture "github.com/tanemahuta/goselfie/internal/test"
)

func updateModeFixture(name string) (string, string) {
	path := filepath.Join(testfixture.PrepareTestDir(name), "update_mode_test.go")
	contents, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	return path, string(contents)
}

func updateModeLineOf(source, text string) int {
	index := strings.Index(source, text)
	Expect(index).NotTo(Equal(-1))
	return strings.Count(source[:index], "\n") + 1
}
