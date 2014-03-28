package emitter_test

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net"
	"os"
)

var _ = Describe("UdpEmitter", func() {

	BeforeEach(func() {
		os.Setenv("BOSH_JOB_NAME", "awesome_job")
		os.Setenv("BOSH_JOB_INSTANCE", "1")
	})

	Context("when the agent is listening", func() {

		var agentListener net.PacketConn
		var testEvent *events.DropsondeStatus

		BeforeEach(func() {
			var err error
			agentListener, err = net.ListenPacket("udp", ":42420")
			Expect(err).To(BeNil())
			testEvent = &events.DropsondeStatus{SentCount: proto.Uint64(1), ErrorCount: proto.Uint64(0)}
		})

		AfterEach(func() {
			agentListener.Close()
		})

		PIt("should send the envelope as a []byte", func(done Done) {
			emitter.Emit(testEvent)
			buffer := make([]byte, 0, 4096)
			_, _, err := agentListener.ReadFrom(buffer)
			Expect(err).To(BeNil())
			var envelope events.Envelope
			err = proto.Unmarshal(buffer, &envelope)
			Expect(err).To(BeNil())
			Expect(envelope.GetEventType()).To(Equal(events.Envelope_DropsondeStatus))
			Expect(envelope.GetDropsondeStatus()).To(Equal(testEvent))
			close(done)
		})
	})
})
