package snapshot

import (
	"errors"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tanemahuta/goselfie/snapshot/codec/content"
	"github.com/tanemahuta/goselfie/utils"
)

var _ = Describe("NewTaken", func() {
	When("snapshot metadata and contents are supplied", func() {
		var taken Taken
		source := utils.Range[int]{Start: 2, EndExcl: 7}

		BeforeEach(func() {
			var err error
			taken, err = NewTaken([]string{"suite", "test"}, content.YAML, source, func(writer io.Writer) error {
				_, err := io.WriteString(writer, "value")
				return err
			})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(taken.Evict)
		})

		It("retains its metadata and contents", func() {
			Expect(taken.Name()).To(Equal([]string{"suite", "test"}))
			Expect(taken.ContentType()).To(Equal(content.YAML))
			Expect(taken.Source()).To(Equal(source))
			Expect(taken.Promoted()).To(BeFalse())
			Expect(taken.Contents()).To(Equal([]byte("value")))
		})
	})

	When("stored data is promoted with a matcher source", func() {
		var taken, promoted Taken
		source := utils.Range[int]{Start: 10, EndExcl: 20}

		BeforeEach(func() {
			var err error
			taken, err = NewTaken([]string{"test"}, content.Unknown, utils.Range[int]{}, func(writer io.Writer) error {
				_, err := io.WriteString(writer, "value")
				return err
			})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(taken.Evict)
			promoted = taken.Promote(source)
		})

		It("marks the same snapshot as promoted and records its source", func() {
			Expect(promoted).To(BeIdenticalTo(taken))
			Expect(promoted.Promoted()).To(BeTrue())
			Expect(promoted.Source()).To(Equal(source))
		})
	})

	When("a temporary snapshot is evicted", func() {
		var taken Taken

		BeforeEach(func() {
			var err error
			taken, err = NewTaken(nil, content.Unknown, utils.Range[int]{}, func(io.Writer) error { return nil })
			Expect(err).NotTo(HaveOccurred())
			Expect(taken.Evict()).To(Succeed())
		})

		It("makes contents unavailable and remains safe to evict again", func() {
			_, err := taken.Contents()
			Expect(err).To(MatchError(ContainSubstring("read temporary snapshot")))
			Expect(taken.Evict()).To(Succeed())
		})
	})

	When("snapshot contents fail to write", func() {
		var takeErr error
		failure := errors.New("failed")

		BeforeEach(func() {
			_, takeErr = NewTaken(nil, content.Unknown, utils.Range[int]{}, func(io.Writer) error { return failure })
		})

		It("returns the write failure", func() { Expect(takeErr).To(MatchError(ContainSubstring(failure.Error()))) })
	})

	When("the contents callback is nil", func() {
		var takeErr error

		BeforeEach(func() { _, takeErr = NewTaken(nil, content.Unknown, utils.Range[int]{}, nil) })

		It("reports the missing callback", func() { Expect(takeErr).To(MatchError(ContainSubstring("contents is nil"))) })
	})
})
