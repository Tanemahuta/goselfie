package lens

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLens(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Lens Suite")
}
