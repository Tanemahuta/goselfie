package gomega

import "github.com/onsi/gomega/types"

// MatcherBuilder is both a Gomega matcher and an immutable snapshot matcher
// configuration. Each configuration method returns an independent builder.
type MatcherBuilder[I any] interface {
	types.GomegaMatcher
	// WithName sets the name for the snapshot in the container.
	WithName(name string) MatcherBuilder[I]
}
