package envelopes_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestEnvelopes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Envelopes Suite")
}
