package example

var _ = Describe("outer", func() {
	Context("inner", func() {
		// selfieonce
		It("target", func() {
			Expect(value).To(MatchSnapshot(projection))
		})
	})
	It("sibling", func() {
		Expect(value).To(MatchSnapshot(projection))
	})
})
