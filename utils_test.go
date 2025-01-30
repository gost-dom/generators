package generators_test

import (
	"testing"

	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	. "github.com/gost-dom/generators"
)

func TestGenerateReceiverName(t *testing.T) {
	expect := gomega.NewWithT(t).Expect
	expect(DefaultReceiverName("HTMLFormElement")).To(Equal("e"))
}
