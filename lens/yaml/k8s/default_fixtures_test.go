package k8s

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	testfixture "github.com/tanemahuta/goselfie/internal/test"
	sigsyaml "sigs.k8s.io/yaml"
)

type resource struct {
	Metadata map[string]any `json:"metadata"`
	Spec     map[string]any `json:"spec"`
}

func loadYAML[T any](name string) T {
	GinkgoHelper()
	var value T
	Expect(sigsyaml.Unmarshal(testfixture.ReadTestFile(name), &value)).To(Succeed())
	return value
}
