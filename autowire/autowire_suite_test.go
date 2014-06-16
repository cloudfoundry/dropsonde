package autowire_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAutowire(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Autowire Suite")
}
