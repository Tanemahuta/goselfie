package failuremessage

import (
	"github.com/onsi/gomega/format"
	"github.com/tanemahuta/goselfie/snapshot/codec/content"
)

//nolint:gochecknoinits // Register this provider when its implementation is imported.
func init() {
	Providers.Register(content.Text, textProvider{})
}

type textProvider struct{}

func (textProvider) FailureMessage(actual []byte, expected []byte) string {
	return format.MessageWithDiff(string(actual), "to match snapshot", string(expected))
}

func (textProvider) NegateFailureMessage(actual []byte, expected []byte) string {
	return format.MessageWithDiff(string(actual), "not to match snapshot", string(expected))
}
