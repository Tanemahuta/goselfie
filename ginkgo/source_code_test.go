package ginkgo

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tanemahuta/goselfie/internal/test"
	"github.com/tanemahuta/goselfie/snapshot"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
	"github.com/tanemahuta/goselfie/utils"
)

var _ = Describe("Source editor", func() {
	When("an updated TODO matcher and a one-shot directive are present", func() {
		var sourcePath string

		BeforeEach(func() {
			directory := test.PrepareTestDir("source_editor")
			sourcePath = filepath.Join(directory, "update_test.go")
		})

		It("cleans only the updated matcher and removes the one-shot directive", func() {
			original, err := os.ReadFile(sourcePath)
			Expect(err).NotTo(HaveOccurred())
			matcherStart := strings.Index(string(original), "MatchSnapshot_TODO(first)")
			taken, err := snapshot.NewTaken([]string{"first"}, content.Text, utils.Range[int]{
				Start:   matcherStart,
				EndExcl: matcherStart + len("MatchSnapshot_TODO"),
			}, func(io.Writer) error { return nil })
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { Expect(taken.Evict()).To(Succeed()) })

			Expect(NewSourceEditor(sourcePath).Cleanup([]snapshot.Taken{taken})).To(Succeed())
			updated, err := os.ReadFile(sourcePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(updated)).To(ContainSubstring("MatchSnapshot(first)"))
			Expect(string(updated)).To(ContainSubstring("MatchSnapshot_TODO(second)"))
			Expect(string(updated)).NotTo(ContainSubstring("selfieonce"))
		})
	})

	When("the source file cannot be read or parsed", func() {
		var readErr, parseErr error

		BeforeEach(func() {
			directory := GinkgoT().TempDir()
			readErr = NewSourceEditor(filepath.Join(directory, "missing_test.go")).Cleanup(nil)
			invalidPath := filepath.Join(directory, "invalid_test.go")
			Expect(os.WriteFile(invalidPath, []byte("package fixture\nfunc {"), 0o600)).To(Succeed())
			parseErr = NewSourceEditor(invalidPath).Cleanup(nil)
		})

		It("reports the failed source operation", func() {
			Expect(readErr).To(MatchError(ContainSubstring("read Ginkgo test source")))
			Expect(parseErr).To(MatchError(ContainSubstring("parse Ginkgo test source")))
		})
	})
})
