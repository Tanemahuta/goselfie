package failuremessage

type testProvider struct{}

func (testProvider) FailureMessage(actual []byte, expected []byte) string {
	return string(actual) + ":" + string(expected)
}

func (testProvider) NegateFailureMessage(actual []byte, expected []byte) string {
	return "not:" + string(actual) + ":" + string(expected)
}
