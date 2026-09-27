package lens

import (
	"fmt"

	yaml "github.com/tanemahuta/goselfie/lens/yaml"
	"github.com/tanemahuta/goselfie/snapshot/content"
	sigsyaml "sigs.k8s.io/yaml"
)

// YAML is a lens serializing the source value to YAML.
type YAML[I any] interface {
	Lens[I, []byte]
	// Transform returns an independent YAML lens with transform appended.
	Transform(transform yaml.Transformer) YAML[I]
}

// ToYAML creates a lens that accepts any input and serializes it to YAML.
func ToYAML() YAML[any] { return ToYAMLOf[any]() }

// ToYAMLOf creates a typed YAML serialization lens.
func ToYAMLOf[I any]() YAML[I] { return &toYAML[I]{} }

// FromYAML creates a lens that deserializes YAML into O.
func FromYAML[O any]() Lens[[]byte, O] {
	return &fromYAML[O]{}
}

type toYAML[I any] struct {
	transform yaml.Transformer
}

func (t *toYAML[I]) String() string {
	return "toYAML"
}
func (t *toYAML[I]) Transform(transform yaml.Transformer) YAML[I] {
	clone := *t
	clone.transform = yaml.Combine(clone.transform, transform)
	return &clone
}

func (t *toYAML[I]) Apply(input I, _ content.Type) ([]byte, content.Type, error) {
	data, err := sigsyaml.Marshal(input)
	if err != nil {
		return nil, content.YAML, fmt.Errorf("serialize value to YAML: %w", err)
	}
	if t.transform == nil {
		return data, content.YAML, nil
	}
	var plain any
	if err := sigsyaml.Unmarshal(data, &plain); err != nil {
		return nil, content.YAML, fmt.Errorf("decode YAML for transforms: %w", err)
	}
	plain, err = yaml.Visit(plain, t.transform)
	if err != nil {
		return nil, content.YAML, fmt.Errorf("apply YAML transforms: %w", err)
	}
	if yaml.IsRemove(plain) {
		plain = nil
	}
	data, err = sigsyaml.Marshal(plain)
	if err != nil {
		return nil, content.YAML, fmt.Errorf("serialize transformed YAML: %w", err)
	}
	return data, content.YAML, nil
}

type fromYAML[O any] struct {
}

func (f *fromYAML[O]) String() string {
	return "fromYAML"
}

func (f *fromYAML[O]) Apply(input []byte, _ content.Type) (O, content.Type, error) {
	var output O
	if err := sigsyaml.Unmarshal(input, &output); err != nil {
		return output, content.Unknown, fmt.Errorf("deserialize YAML: %w", err)
	}
	return output, content.Unknown, nil
}
