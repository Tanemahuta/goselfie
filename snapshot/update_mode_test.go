package snapshot

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Update mode coercion", func() {
	When("two update modes are combined", func() {
		It("keeps the more restrictive mode", func() {
			Expect(CoerceMode(UpdateModeAlways, UpdateModeOnce)).To(Equal(UpdateModeAlways))
			Expect(CoerceMode(UpdateModeMissing, UpdateModeNever)).To(Equal(UpdateModeNever))
			Expect(CoerceMode(UpdateModeOnce, UpdateModeAlways)).To(Equal(UpdateModeAlways))
		})
	})
})
