// Package goselfie provides the snapshot matcher constructor for Gomega assertions.
package goselfie

import (
	"github.com/tanemahuta/goselfie/gomega"
	"github.com/tanemahuta/goselfie/lens"
)

// MatchSnapshot creates a configurable matcher for the next snapshot in the current Ginkgo spec.
func MatchSnapshot[I any](projection lens.Lens[I, []byte]) gomega.MatcherBuilder[I] {
	return gomega.NewMatcherBuilder(projection)
}

// MatchSnapshot_TODO creates a matcher which always updates this snapshot.
// After a successful spec, goselfie rewrites this call to MatchSnapshot.
//
//nolint:revive // The public name preserves Selfie's `_TODO` update suffix.
func MatchSnapshot_TODO[I any](projection lens.Lens[I, []byte]) gomega.MatcherBuilder[I] {
	return gomega.NewUpdatingMatcherBuilder(projection)
}
