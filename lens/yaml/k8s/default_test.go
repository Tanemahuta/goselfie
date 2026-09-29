package k8s

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tanemahuta/goselfie/lens"
	"github.com/tanemahuta/goselfie/lens/yaml"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
	sigsyaml "sigs.k8s.io/yaml"
)

var _ = Describe("Default Kubernetes transformer", func() {
	When("resources contain runtime metadata", func() {
		var result, expected any

		BeforeEach(func() {
			metadata := func(name string) map[string]any {
				return map[string]any{
					"name": name, "namespace": "team-a", "uid": "generated", "resourceVersion": "42",
					"generation": 3, "creationTimestamp": "2026-01-01T00:00:00Z",
					"deletionTimestamp": "2026-01-02T00:00:00Z", "deletionGracePeriodSeconds": 30,
					"managedFields": []any{"manager"}, "selfLink": "/api/resource",
					"labels":          map[string]any{"app": "example"},
					"annotations":     map[string]any{"note": "keep"},
					"ownerReferences": []any{map[string]any{"name": "owner", "uid": "owner-uid"}},
				}
			}
			value := map[string]any{
				"metadata": metadata("root"),
				"spec":     map[string]any{"uid": "stable", "generation": 7},
				"items":    []any{map[string]any{"metadata": metadata("child")}},
			}
			expectedMetadata := func(name string) map[string]any {
				return map[string]any{
					"name": name, "namespace": "team-a", "labels": map[string]any{"app": "example"},
					"annotations":     map[string]any{"note": "keep"},
					"ownerReferences": []any{map[string]any{"name": "owner"}},
				}
			}
			want := map[string]any{
				"metadata": expectedMetadata("root"),
				"spec":     map[string]any{"uid": "stable", "generation": 7},
				"items":    []any{map[string]any{"metadata": expectedMetadata("child")}},
			}
			var err error
			result, err = yaml.Visit(value, DefaultTransformer())
			Expect(err).NotTo(HaveOccurred())
			expected = want
		})

		It("removes runtime metadata at every resource level", func() { Expect(result).To(Equal(expected)) })
	})

	When("a typed resource contains only runtime metadata", func() {
		type resource struct {
			Metadata map[string]any `json:"metadata"`
			Spec     map[string]any `json:"spec"`
		}
		var plain any

		BeforeEach(func() {
			value := resource{
				Metadata: map[string]any{"uid": "dynamic", "generation": 2},
				Spec:     map[string]any{"uid": "stable"},
			}
			data, _, err := lens.ToYAML().Transform(DefaultTransformer()).Apply(value, content.Unknown)
			Expect(err).NotTo(HaveOccurred())
			Expect(sigsyaml.Unmarshal(data, &plain)).To(Succeed())
		})

		It("removes metadata emptied by normalization", func() {
			Expect(plain).To(Equal(map[string]any{"spec": map[string]any{"uid": "stable"}}))
		})
	})
})
