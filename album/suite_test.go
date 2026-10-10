package album

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAlbum(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Album Suite")
}
