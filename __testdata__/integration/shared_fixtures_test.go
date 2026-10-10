package integration

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tanemahuta/goselfie/lens"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
)

func loadYAML(name string) any {
	GinkgoHelper()
	data, err := os.ReadFile(filepath.Join("__testdata__", filepath.FromSlash(name)))
	Expect(err).NotTo(HaveOccurred())
	value, _, err := lens.FromYAML[any]().Apply(data, content.YAML)
	Expect(err).NotTo(HaveOccurred())
	return value
}
