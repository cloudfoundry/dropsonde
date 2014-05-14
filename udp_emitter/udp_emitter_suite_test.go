package udp_emitter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestUdpemitter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Udpemitter Suite")
}
