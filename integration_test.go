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
	file          string
	container     string
	expectedValue string
	directive     string
	removed       bool
}

var _ = Describe("Public snapshot integration", func() {
	directiveScenarios := []updateDirectiveScenario{
		{file: "snapshot_test.go", container: "file selfieonce", expectedValue: "example", directive: "selfieonce", removed: true},
		{file: "file_selfiewrite_test.go", container: "file SELFIEWRITE", expectedValue: "file-always", directive: "SELFIEWRITE"},
		{file: "container_selfieonce_test.go", container: "container selfieonce", expectedValue: "container-once", directive: "selfieonce", removed: true},
		{file: "container_selfiewrite_test.go", container: "container SELFIEWRITE", expectedValue: "container-always", directive: "SELFIEWRITE"},
	}

	When("file and container directives request updates", func() {
		var module string

		BeforeEach(func() {
			module = prepareTestGoProject()
			runIntegrationTests(module)
		})

		It("persists each snapshot and cleans one-shot directives", func() {
			for _, scenario := range directiveScenarios {
				assertIntegratedSnapshot(module, scenario)
			}
		})

		When("the updated module suite runs again", func() {
			BeforeEach(func() { runIntegrationTests(module) })

			It("matches all persisted snapshots", func() {
				for _, scenario := range directiveScenarios {
					assertIntegratedSnapshot(module, scenario)
				}
			})
		})
	})

	When("CI blocks writes for missing snapshots", func() {
		var module string
		var runErr error

		BeforeEach(func() {
			module = prepareTestGoProject()
			Expect(os.RemoveAll(filepath.Join(module, album.SnapshotsDirName))).To(Succeed())
			_, runErr = executeIntegrationTests(module, "CI=true")
		})

		It("fails without creating the snapshots directory", func() {
			Expect(runErr).To(HaveOccurred())
			_, err := os.Stat(filepath.Join(module, album.SnapshotsDirName))
			Expect(os.IsNotExist(err)).To(BeTrue())
		})
	})

	When("a test file contains multiple Its", func() {
		var module string

		BeforeEach(func() {
			module = prepareTestGoProject()
			runIntegrationTests(module)
		})

		It("stores every It snapshot under its container path", func() {
			testFile := filepath.Join(module, "environment_test.go")
			stored, err := album.OpenAlbum(testFile)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(stored.Close)
			Expect(stored.AllSnapshots()).To(HaveLen(2))
			first := stored.SnapshotByName([]string{"environment", "updates existing", "value"})
			second := stored.SnapshotByName([]string{"environment", "writes another container", "value"})
			Expect(first).NotTo(BeNil())
			Expect(second).NotTo(BeNil())
			Expect(first.Contents()).To(Equal([]byte("name: environment-original\n")))
			Expect(second.Contents()).To(Equal([]byte("name: environment-second\n")))
		})
	})

	When("a persisted snapshot value changes", func() {
		var module string
		var albumFile string
		var original []byte

		BeforeEach(func() {
			module = prepareTestGoProject()
			runIntegrationTests(module)
			albumFile = filepath.Join(module, album.SnapshotsDirName, "environment_test.ss")
			var err error
			original, err = os.ReadFile(albumFile)
			Expect(err).NotTo(HaveOccurred())
			path := filepath.Join(module, "environment_test.go")
			contents, err := os.ReadFile(path)
			Expect(err).NotTo(HaveOccurred())
			contents = []byte(strings.ReplaceAll(string(contents), "environment-original", "environment-ci-change"))
			Expect(os.WriteFile(path, contents, 0o600)).To(Succeed())
		})

		When("the changed suite runs with CI enabled", func() {
			var runErr error

			BeforeEach(func() { _, runErr = executeIntegrationTests(module, "CI=true") })

			It("fails and preserves the existing album", func() {
				Expect(runErr).To(HaveOccurred())
				Expect(os.ReadFile(albumFile)).To(Equal(original))
			})
		})
	})

	When("snapshot updates are enabled through the environment", func() {
		var module string
		var stored album.Album

		BeforeEach(func() {
			module = prepareTestGoProject()
			runIntegrationTests(module)
			path := filepath.Join(module, "environment_test.go")
			contents, err := os.ReadFile(path)
			Expect(err).NotTo(HaveOccurred())
			contents = []byte(strings.ReplaceAll(string(contents), "environment-original", "environment-updated"))
			Expect(os.WriteFile(path, contents, 0o600)).To(Succeed())
			output, err := executeIntegrationTests(module, "UPDATE_SNAPSHOTS=true")
			Expect(err).NotTo(HaveOccurred(), strings.TrimSpace(string(output)))
		})
		AfterEach(func() {
			if stored != nil {
				Expect(stored.Close()).To(Succeed())
			}
		})

		It("replaces the existing snapshot", func() {
			var err error
			stored, err = album.OpenAlbum(filepath.Join(module, "environment_test.go"))
			Expect(err).NotTo(HaveOccurred())
			actual := stored.SnapshotByName([]string{"environment", "updates existing", "value"})
			Expect(actual).NotTo(BeNil())
			Expect(actual.Contents()).To(Equal([]byte("name: environment-updated\n")))
		})
	})
})

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
	Expect(actual.Contents()).To(Equal([]byte("name: " + scenario.expectedValue + "\n")))
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
