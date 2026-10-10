package ginkgo

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/tanemahuta/goselfie/internal/test"
)

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
