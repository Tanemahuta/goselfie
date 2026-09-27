// Package gomega implements goselfie's typed, fluent snapshot matcher.
//
// MatcherBuilder extends types.GomegaMatcher, so a configured builder can be
// passed directly to Expect(value).To(...). Builder methods are immutable and
// may be reused to derive independent matcher configurations.
package gomega
