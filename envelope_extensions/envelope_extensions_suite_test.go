package envelope_extensions_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestEnvelopeExtensions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EnvelopeExtensions Suite")
}
