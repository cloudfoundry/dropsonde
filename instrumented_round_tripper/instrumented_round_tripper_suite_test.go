package instrumented_round_tripper_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestInstrumentedRoundTripper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "InstrumentedRoundTripper Suite")
}
