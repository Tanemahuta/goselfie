package ginkgo

import (
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tanemahuta/goselfie/snapshot"
)

var _ = Describe("CurrentTestContext", func() {
	When("a context is requested from a Ginkgo spec", func() {
		// SELFIEWRITE
		It("reuses the file album and applies the container update mode", func() {
			first, err := CurrentTestContext()
			Expect(err).NotTo(HaveOccurred())
			second, err := CurrentTestContext()
			Expect(err).NotTo(HaveOccurred())

			_, sourcePath, _, ok := runtime.Caller(0)
			Expect(ok).To(BeTrue())
			absolutePath, err := filepath.Abs(sourcePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(first.TestFilePath()).To(Equal(absolutePath))
			Expect(first.Album().TestFilePath()).To(Equal(absolutePath))
			Expect(first.Album()).To(BeIdenticalTo(second.Album()))
			Expect(first.Album().FilePath()).To(HaveSuffix(filepath.Join("__snapshots__", "test_context_integration_test.ss")))
			Expect(first.UpdateMode()).To(Equal(snapshot.UpdateModeAlways))
			Expect(first.Close()).To(Succeed())
		})
	})
})
