package content

import (
	"bytes"
	"io"

	yamlv3 "gopkg.in/yaml.v3"
)

const yamlPriority = 0

func init() { Matchers.Register(YAMLContent{}) }

// YAMLContent matches valid YAML streams and stores them verbatim.
type YAMLContent struct{}

// ContentType returns YAML.
func (YAMLContent) ContentType() Type { return YAML }

// Priority makes YAML detection run before text and binary detection.
func (YAMLContent) Priority() int { return yamlPriority }

// Matches accepts non-empty streams that decode completely as YAML.
func (YAMLContent) Matches(contents []byte) bool {
	decoder := yamlv3.NewDecoder(bytes.NewReader(contents))
	documents := 0
	for {
		var document any
		err := decoder.Decode(&document)
		if err == io.EOF {
			return documents > 0
		}
		if err != nil {
			return false
		}
		documents++
	}
}

// Encode returns YAML bytes verbatim as a string.
func (YAMLContent) Encode(contents []byte) (string, error) { return string(contents), nil }
