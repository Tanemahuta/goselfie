package goselfie

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGoselfie(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Goselfie Suite")
}
