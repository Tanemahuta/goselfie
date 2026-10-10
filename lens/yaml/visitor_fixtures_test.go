package yaml

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	testfixture "github.com/tanemahuta/goselfie/internal/test"
	sigsyaml "sigs.k8s.io/yaml"
)

func loadYAML(name string) any {
	GinkgoHelper()
	var value any
	Expect(sigsyaml.Unmarshal(testfixture.ReadTestFile(name), &value)).To(Succeed())
	return value
}

func visitYAML(name string, transforms ...Transformer) any {
	GinkgoHelper()
	result, err := Visit(loadYAML(name), transforms...)
	Expect(err).NotTo(HaveOccurred())
	return result
}

func recordingTransformer(calls *[]string, name string, priority int) Transformer {
	return NewTransformer(nil, func(value any) (any, error) {
		*calls = append(*calls, name)
		return value, nil
	}).WithPriority(priority)
}

func recordingPathTransformer(calls *[]string, target string, replacement any) Transformer {
	return NewTransformer(func(path []string) bool {
		name := formatPath(path)
		*calls = append(*calls, name)
		return name == target
	}, func(any) (any, error) { return replacement, nil })
}
