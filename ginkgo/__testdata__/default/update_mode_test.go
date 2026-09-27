package example

var _ = It("test", func() {
	Expect(value).To(MatchSnapshot(projection))
})
