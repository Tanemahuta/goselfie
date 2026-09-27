package example

var _ = It("test", func() {
	Expect(value).To(
		// SELFIEWRITE
		MatchSnapshot(projection),
	)
})
