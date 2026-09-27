package example

var _ = It("test", func() {
	Expect(value).To(
		// selfieonce
		MatchSnapshot(projection),
	)
})
