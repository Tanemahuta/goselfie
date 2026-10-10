package goselfie

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tanemahuta/goselfie/album"
	testfixture "github.com/tanemahuta/goselfie/internal/test"
)

var _ = Describe("Public snapshot integration", func() {
	directiveScenarios := []updateDirectiveScenario{
		{file: "snapshot_test.go", container: "file selfieonce", expected: "expected/file_selfieonce.yaml", directive: "selfieonce", removed: true},
		{file: "file_selfiewrite_test.go", container: "file SELFIEWRITE", expected: "expected/file_selfiewrite.yaml", directive: "SELFIEWRITE"},
		{file: "container_selfieonce_test.go", container: "container selfieonce", expected: "expected/container_selfieonce.yaml", directive: "selfieonce", removed: true},
		{file: "container_selfiewrite_test.go", container: "container SELFIEWRITE", expected: "expected/container_selfiewrite.yaml", directive: "SELFIEWRITE"},
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
			stored, err := album.OpenAlbum(filepath.Join(module, "environment_test.go"))
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(stored.Close)
			Expect(stored.AllSnapshots()).To(HaveLen(2))
			Expect(storedSnapshot(module, "environment_test.go", "environment", "updates existing", "value")).
				To(Equal(testfixture.ReadTestFile("expected/environment_updates_existing.yaml")))
			Expect(storedSnapshot(module, "environment_test.go", "environment", "writes another container", "value")).
				To(Equal(testfixture.ReadTestFile("expected/environment_another_container.yaml")))
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
			replaceModuleTestData(module, "environment/updates_existing.yaml", "changes/environment_ci_change.yaml")
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

		BeforeEach(func() {
			module = prepareTestGoProject()
			runIntegrationTests(module)
			replaceModuleTestData(module, "environment/updates_existing.yaml", "changes/environment_updated.yaml")
			output, err := executeIntegrationTests(module, "UPDATE_SNAPSHOTS=true")
			Expect(err).NotTo(HaveOccurred(), strings.TrimSpace(string(output)))
		})

		It("replaces the existing snapshot", func() {
			Expect(storedSnapshot(module, "environment_test.go", "environment", "updates existing", "value")).
				To(Equal(testfixture.ReadTestFile("expected/environment_updated.yaml")))
		})
	})
})
