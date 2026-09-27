package ginkgo

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGinkgoContext(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ginkgo Context Suite")
}
