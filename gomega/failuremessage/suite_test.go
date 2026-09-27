package failuremessage

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFailureMessage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Failure Message Suite")
}
