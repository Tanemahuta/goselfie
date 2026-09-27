package ginkgo

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tanemahuta/goselfie/album"
	"github.com/tanemahuta/goselfie/internal/test"
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
				name     string
				contents string
			}{
				{name: "first spec", contents: "name: first\n"},
				{name: "second spec", contents: "name: second\n"},
			} {
				actual := stored.SnapshotByName([]string{"shared test file", expected.name, "#1"})
				Expect(actual).NotTo(BeNil())
				contents, err := actual.Contents()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal(expected.contents))
			}
		})
	})
})

func prepareContextCacheModule() string {
	module := test.PrepareTestDir("context_cache")
	_, sourceFile, _, ok := runtime.Caller(0)
	Expect(ok).To(BeTrue())
	moduleRoot := filepath.Dir(filepath.Dir(sourceFile))

	goModPath := filepath.Join(module, "go.mod")
	goMod, err := os.ReadFile(goModPath)
	Expect(err).NotTo(HaveOccurred())
	goMod = []byte(strings.ReplaceAll(string(goMod), "GOSELFIE_ROOT", moduleRoot))
	Expect(os.WriteFile(goModPath, goMod, 0o600)).To(Succeed())

	goSum, err := os.ReadFile(filepath.Join(moduleRoot, "go.sum"))
	Expect(err).NotTo(HaveOccurred())
	Expect(os.WriteFile(filepath.Join(module, "go.sum"), goSum, 0o600)).To(Succeed())
	return module
}

func contextCacheEnvironment() []string {
	blocked := map[string]bool{"CI": true, "UPDATE_SNAPSHOTS": true, "GOWORK": true}
	result := make([]string, 0, len(os.Environ())+3)
	for _, value := range os.Environ() {
		name, _, _ := strings.Cut(value, "=")
		if !blocked[name] {
			result = append(result, value)
		}
	}
	return append(result, "CI=false", "UPDATE_SNAPSHOTS=false", "GOWORK=off")
}
