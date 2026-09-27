package ginkgo

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	testfixture "github.com/tanemahuta/goselfie/internal/test"
	"github.com/tanemahuta/goselfie/snapshot"
	"github.com/tanemahuta/goselfie/utils"
)

var _ = Describe("Update mode detection", func() {
	When("source has no update directives", func() {
		var fileMode, containerMode, matcherMode snapshot.UpdateMode

		BeforeEach(func() {
			path, source := updateModeFixture("default")
			var err error
			fileMode, err = detectFileUpdateMode(path)
			Expect(err).NotTo(HaveOccurred())
			containerMode, err = detectContainerUpdateMode(path, 3)
			Expect(err).NotTo(HaveOccurred())
			start := strings.Index(source, "MatchSnapshot(projection)")
			matcherMode, err = UpdateModeForSource(path, utils.Range[int]{Start: start, EndExcl: start + len("MatchSnapshot(projection)")})
			Expect(err).NotTo(HaveOccurred())
		})

		It("defaults each scope to missing-only updates", func() {
			Expect(fileMode).To(Equal(snapshot.UpdateModeMissing))
			Expect(containerMode).To(Equal(snapshot.UpdateModeMissing))
			Expect(matcherMode).To(Equal(snapshot.UpdateModeMissing))
		})
	})

	When("a file contains both SELFIEWRITE and selfieonce", func() {
		var mode snapshot.UpdateMode

		BeforeEach(func() {
			path, _ := updateModeFixture("file_priority")
			var err error
			mode, err = detectFileUpdateMode(path)
			Expect(err).NotTo(HaveOccurred())
		})

		It("gives SELFIEWRITE precedence", func() { Expect(mode).To(Equal(snapshot.UpdateModeAlways)) })
	})

	When("a selfieonce mode has already been detected", func() {
		var path string
		var mode snapshot.UpdateMode

		BeforeEach(func() {
			path, _ = updateModeFixture("file_selfieonce")
			var err error
			mode, err = detectFileUpdateMode(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(mode).To(Equal(snapshot.UpdateModeOnce))
			contents, err := os.ReadFile(path)
			Expect(err).NotTo(HaveOccurred())
			contents = []byte(strings.ReplaceAll(string(contents), "// selfieonce", ""))
			Expect(os.WriteFile(path, contents, 0o600)).To(Succeed())
		})

		It("retains the detected mode after the comment is removed", func() {
			actual, err := detectFileUpdateMode(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(mode))
		})
	})

	for _, scenario := range []struct {
		name, fixture string
		expected      snapshot.UpdateMode
	}{
		{"file SELFIEWRITE", "file_selfiewrite", snapshot.UpdateModeAlways},
		{"file selfieonce", "file_selfieonce", snapshot.UpdateModeOnce},
	} {
		scenario := scenario
		When(scenario.name+" is present", func() {
			var mode snapshot.UpdateMode

			BeforeEach(func() {
				path, _ := updateModeFixture(scenario.fixture)
				var err error
				mode, err = detectFileUpdateMode(path)
				Expect(err).NotTo(HaveOccurred())
			})

			It("selects the matching file mode", func() { Expect(mode).To(Equal(scenario.expected)) })
		})
	}

	for _, scenario := range []struct {
		name, fixture string
		expected      snapshot.UpdateMode
	}{
		{"container SELFIEWRITE", "container_selfiewrite", snapshot.UpdateModeAlways},
		{"container selfieonce", "container_selfieonce", snapshot.UpdateModeOnce},
	} {
		scenario := scenario
		When(scenario.name+" is present", func() {
			var path, source string
			var containerMode, fileMode, matcherMode snapshot.UpdateMode

			BeforeEach(func() {
				path, source = updateModeFixture(scenario.fixture)
				var err error
				containerMode, err = detectContainerUpdateMode(path, 3)
				Expect(err).NotTo(HaveOccurred())
				fileMode, err = detectFileUpdateMode(path)
				Expect(err).NotTo(HaveOccurred())
				start := strings.Index(source, "MatchSnapshot(projection)")
				matcherMode, err = UpdateModeForSource(path, utils.Range[int]{Start: start, EndExcl: start + len("MatchSnapshot(projection)")})
				Expect(err).NotTo(HaveOccurred())
			})

			It("applies only to the container scope", func() {
				Expect(containerMode).To(Equal(scenario.expected))
				Expect(fileMode).To(Equal(snapshot.UpdateModeMissing))
				Expect(matcherMode).To(Equal(snapshot.UpdateModeMissing))
			})

		})
	}

	for _, scenario := range []struct {
		name, fixture string
		expected      snapshot.UpdateMode
	}{
		{"matcher SELFIEWRITE", "matcher_selfiewrite", snapshot.UpdateModeAlways},
		{"matcher selfieonce", "matcher_selfieonce", snapshot.UpdateModeOnce},
	} {
		scenario := scenario
		When(scenario.name+" is present", func() {
			var matcherMode, containerMode snapshot.UpdateMode

			BeforeEach(func() {
				path, source := updateModeFixture(scenario.fixture)
				start := strings.Index(source, "MatchSnapshot(projection)")
				var err error
				matcherMode, err = UpdateModeForSource(path, utils.Range[int]{Start: start, EndExcl: start + len("MatchSnapshot(projection)")})
				Expect(err).NotTo(HaveOccurred())
				containerMode, err = detectContainerUpdateMode(path, 3)
				Expect(err).NotTo(HaveOccurred())
			})

			It("applies only to the matcher scope", func() {
				Expect(matcherMode).To(Equal(scenario.expected))
				Expect(containerMode).To(Equal(snapshot.UpdateModeMissing))
			})
		})
	}

	When("a nested context has a local selfieonce directive", func() {
		var outerMode, innerMode, targetMode, siblingMode snapshot.UpdateMode

		BeforeEach(func() {
			path, source := updateModeFixture("nested_container")
			lineOf := func(text string) int {
				index := strings.Index(source, text)
				Expect(index).NotTo(Equal(-1))
				return strings.Count(source[:index], "\n") + 1
			}
			outerLine := lineOf(`Describe("outer"`)
			innerLine := lineOf(`Context("inner"`)
			targetLine := lineOf(`It("target"`)
			siblingLine := lineOf(`It("sibling"`)
			var err error
			outerMode, err = detectContainerUpdateMode(path, outerLine)
			Expect(err).NotTo(HaveOccurred())
			innerMode, err = detectContainerUpdateMode(path, innerLine)
			Expect(err).NotTo(HaveOccurred())
			targetMode, err = detectContainerUpdateMode(path, targetLine)
			Expect(err).NotTo(HaveOccurred())
			siblingMode, err = detectContainerUpdateMode(path, siblingLine)
			Expect(err).NotTo(HaveOccurred())
		})

		It("scopes the directive to the nested container", func() {
			Expect(outerMode).To(Equal(snapshot.UpdateModeMissing))
			Expect(innerMode).To(Equal(snapshot.UpdateModeOnce))
			Expect(targetMode).To(Equal(snapshot.UpdateModeMissing))
			Expect(siblingMode).To(Equal(snapshot.UpdateModeMissing))
		})
	})
})

func updateModeFixture(name string) (string, string) {
	path := filepath.Join(testfixture.PrepareTestDir(name), "update_mode_test.go")
	contents, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	return path, string(contents)
}
