package integration

import (
	"github.com/tanemahuta/goselfie/lens"
	"github.com/tanemahuta/goselfie/lens/yaml"
	"github.com/tanemahuta/goselfie/lens/yaml/k8s"
)

func normalized() lens.YAML[any] {
	return lens.ToYAML().Transform(yaml.Combine(
		k8s.DefaultTransformer(),
		yaml.IgnoreOrder("items"),
		yaml.IgnoreOrder("items.*.ports"),
		yaml.IgnoreOrder("**.matrix"),
		yaml.IgnoreOrder("**.matrix.*"),
		yaml.Mask("items.*.token"),
		yaml.Remove("items.*.uid"),
	))
}
