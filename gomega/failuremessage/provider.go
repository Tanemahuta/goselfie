// Package failuremessage formats snapshot matcher failures by content type.
package failuremessage

// Provider formats positive and negated snapshot matcher failures.
type Provider interface {
	FailureMessage(actual []byte, expected []byte) string
	NegateFailureMessage(actual []byte, expected []byte) string
}
