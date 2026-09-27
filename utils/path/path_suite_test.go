package path

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPath(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Path Matcher Suite")
}
