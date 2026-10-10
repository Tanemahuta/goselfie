package goselfie

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tanemahuta/goselfie/album"
	testfixture "github.com/tanemahuta/goselfie/internal/test"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
)

type updateDirectiveScenario struct {
	file      string
	container string
	expected  string
	directive string
	removed   bool
}

func assertIntegratedSnapshot(module string, scenario updateDirectiveScenario) {
	testFile := filepath.Join(module, scenario.file)
	updated, err := os.ReadFile(testFile)
	Expect(err).NotTo(HaveOccurred())
	Expect(string(updated)).NotTo(ContainSubstring("MatchSnapshot_TODO"))
	Expect(string(updated)).To(ContainSubstring("MatchSnapshot(lens.ToYAML())"))
	if scenario.removed {
		Expect(string(updated)).NotTo(ContainSubstring("// " + scenario.directive))
	} else {
		Expect(string(updated)).To(ContainSubstring("// " + scenario.directive))
	}

	stored, err := album.OpenAlbum(testFile)
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(stored.Close)
	actual := stored.SnapshotByName([]string{scenario.container, "writes", "value"})
	Expect(actual).NotTo(BeNil())
	Expect(actual.Contents()).To(Equal(testfixture.ReadTestFile(scenario.expected)))
	Expect(actual.ContentType()).To(Equal(content.YAML))
}

func prepareTestGoProject() string {
	result := testfixture.PrepareTestDir("integration")
	root := moduleRoot()
	goMod, err := os.ReadFile(filepath.Join(result, "go.mod"))
	Expect(err).NotTo(HaveOccurred())
	goMod = []byte(strings.ReplaceAll(string(goMod), "GOSELFIE_ROOT", root))
	Expect(os.WriteFile(filepath.Join(result, "go.mod"), goMod, 0o600)).To(Succeed())
	checksums, err := os.ReadFile(filepath.Join(root, "go.sum"))
	Expect(err).NotTo(HaveOccurred())
	Expect(os.WriteFile(filepath.Join(result, "go.sum"), checksums, 0o600)).To(Succeed())
	return result
}

func moduleRoot() string {
	_, file, _, ok := runtime.Caller(0)
	Expect(ok).To(BeTrue())
	return filepath.Dir(file)
}

func runIntegrationTests(module string) {
	output, err := executeIntegrationTests(module)
	Expect(err).NotTo(HaveOccurred(), strings.TrimSpace(string(output)))
}

func executeIntegrationTests(module string, environment ...string) ([]byte, error) {
	command := exec.Command("go", "test", "-mod=mod", "-count=1", "./...")
	command.Dir = module
	command.Env = integrationEnvironment(environment...)
	return command.CombinedOutput()
}

func integrationEnvironment(overrides ...string) []string {
	blocked := map[string]bool{"CI": true, "UPDATE_SNAPSHOTS": true, "GOWORK": true}
	result := make([]string, 0, len(os.Environ())+len(overrides)+3)
	for _, value := range os.Environ() {
		name, _, _ := strings.Cut(value, "=")
		if !blocked[name] {
			result = append(result, value)
		}
	}
	result = append(result, "CI=false", "UPDATE_SNAPSHOTS=false", "GOWORK=off")
	return append(result, overrides...)
}

// replaceModuleTestData overwrites a file in the prepared module's __testdata__
// with a file from this package's __testdata__.
func replaceModuleTestData(module, target, source string) {
	GinkgoHelper()
	path := filepath.Join(module, "__testdata__", filepath.FromSlash(target))
	Expect(os.WriteFile(path, testfixture.ReadTestFile(source), 0o600)).To(Succeed())
}

func storedSnapshot(module, file string, name ...string) []byte {
	GinkgoHelper()
	stored, err := album.OpenAlbum(filepath.Join(module, file))
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(stored.Close)
	snapshot := stored.SnapshotByName(name)
	Expect(snapshot).NotTo(BeNil())
	data, err := snapshot.Contents()
	Expect(err).NotTo(HaveOccurred())
	return data
}
