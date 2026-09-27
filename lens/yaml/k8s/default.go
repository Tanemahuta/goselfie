// Package k8s provides YAML transforms for Kubernetes resource snapshots.
package k8s

import "github.com/tanemahuta/goselfie/lens/yaml"

// DefaultTransformer removes metadata that Kubernetes changes at runtime.
// It matches both a resource at the root and resources nested in lists or maps.
// Names, labels, annotations, and status remain available for assertions.
func DefaultTransformer() yaml.Transformer {
	return yaml.Combine(
		yaml.Remove("**.metadata.uid"),
		yaml.Remove("**.metadata.resourceVersion"),
		yaml.Remove("**.metadata.generation"),
		yaml.Remove("**.metadata.creationTimestamp"),
		yaml.Remove("**.metadata.deletionTimestamp"),
		yaml.Remove("**.metadata.deletionGracePeriodSeconds"),
		yaml.Remove("**.metadata.managedFields"),
		yaml.Remove("**.metadata.selfLink"),
		yaml.Remove("**.metadata.ownerReferences.*.uid"),
	)
}
