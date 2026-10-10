package lens

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	testfixture "github.com/tanemahuta/goselfie/internal/test"
	yaml "github.com/tanemahuta/goselfie/lens/yaml"
	sigsyaml "sigs.k8s.io/yaml"
)

type yamlFixture struct {
	Metadata map[string]any `json:"metadata" yaml:"metadata"`
	Spec     map[string]any `json:"spec" yaml:"spec"`
}

func plainYAML(data []byte) any {
	GinkgoHelper()
	var value any
	Expect(sigsyaml.Unmarshal(data, &value)).To(Succeed())
	return value
}

func loadYAML[T any](name string) T {
	GinkgoHelper()
	var value T
	Expect(sigsyaml.Unmarshal(testfixture.ReadTestFile(name), &value)).To(Succeed())
	return value
}

func recordingTransformer(calls *[]string, name string, priority int) yaml.Transformer {
	return yaml.NewTransformer(nil, func(value any) (any, error) {
		*calls = append(*calls, name)
		return value, nil
	}).WithPriority(priority)
}
