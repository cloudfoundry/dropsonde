package dropsonde_marshaller_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestUnmarshaller(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dropsonde Marshaller Suite")
}
