package failuremessage

import (
	"github.com/onsi/gomega/format"
	"github.com/tanemahuta/goselfie/snapshot/content"
)

func init() {
	Providers.Register(content.YAML, yamlProvider{})
}

type yamlProvider struct{}

func (yamlProvider) FailureMessage(actual []byte, expected []byte) string {
	return format.MessageWithDiff(string(actual), "to match YAML snapshot", string(expected))
}

func (yamlProvider) NegateFailureMessage(actual []byte, expected []byte) string {
	return format.MessageWithDiff(string(actual), "not to match YAML snapshot", string(expected))
}
