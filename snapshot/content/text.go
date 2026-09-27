package content

import (
	"unicode"
	"unicode/utf8"
)

const textPriority = 1

//nolint:gochecknoinits // Register the text content matcher when this file's package loads.
func init() { Matchers.Register(TextContent{}) }

// TextContent matches valid printable UTF-8 text and stores it verbatim.
type TextContent struct{}

// ContentType returns Text.
func (TextContent) ContentType() Type { return Text }

// Priority makes text detection run after YAML and before binary detection.
func (TextContent) Priority() int { return textPriority }

// Matches reports whether contents are valid UTF-8 without non-printable characters.
func (TextContent) Matches(contents []byte) bool {
	if !utf8.Valid(contents) {
		return false
	}
	for len(contents) > 0 {
		value, size := utf8.DecodeRune(contents)
		if !unicode.IsPrint(value) && value != '\n' && value != '\r' && value != '\t' {
			return false
		}
		contents = contents[size:]
	}
	return true
}

// Encode returns text bytes verbatim as a string.
func (TextContent) Encode(contents []byte) (string, error) { return string(contents), nil }
