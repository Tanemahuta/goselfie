package ginkgo

import (
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tanemahuta/goselfie/album"
	testfixture "github.com/tanemahuta/goselfie/internal/test"
)

var _ = Describe("test context cache integration", func() {
	When("multiple Ginkgo specs share one test file", func() {
		var module string

		BeforeEach(func() {
			module = prepareContextCacheModule()
			command := exec.Command("go", "test", "-mod=mod", "-count=1", "./...")
			command.Dir = module
			command.Env = contextCacheEnvironment()
			output, err := command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), strings.TrimSpace(string(output)))
		})

		It("writes both specs to that file's album when the suite closes", func() {
			stored, err := album.OpenAlbum(filepath.Join(module, "cache_test.go"))
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(stored.Close)
			Expect(stored.AllSnapshots()).To(HaveLen(2))

			for _, expected := range []struct {
				name    string
				fixture string
			}{
				{name: "first spec", fixture: "context_cache_expected/first.yaml"},
				{name: "second spec", fixture: "context_cache_expected/second.yaml"},
			} {
				actual := stored.SnapshotByName([]string{"shared test file", expected.name, "#1"})
				Expect(actual).NotTo(BeNil())
				contents, err := actual.Contents()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal(string(testfixture.ReadTestFile(expected.fixture))))
			}
		})
	})
})
