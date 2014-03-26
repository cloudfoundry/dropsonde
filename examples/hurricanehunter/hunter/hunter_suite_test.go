package hunter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHunter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hunter Suite")
}
