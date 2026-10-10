package content

type testMatcher struct {
	contentType Type
	priority    int
	matches     bool
}

func (matcher testMatcher) ContentType() Type   { return matcher.contentType }
func (matcher testMatcher) Priority() int       { return matcher.priority }
func (matcher testMatcher) Matches([]byte) bool { return matcher.matches }
func (testMatcher) Encode(contents []byte) (string, error) {
	return string(contents), nil
}
