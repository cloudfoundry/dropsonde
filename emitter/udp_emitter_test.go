package emitter_test

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net"
)

var _ = Describe("UdpEmitter", func() {

	var jobName = "testInstrumentedEmitter"
	var jobIndex int32 = 42
	var origin = events.NewOrigin(jobName, jobIndex)
	var testEvent = events.NewTestEvent(43)
	var remoteAddr = "localhost:12345"

	Describe("Close()", func() {
		It("closes the UDP connection", func() {

			udpEmitter, _ := emitter.NewUdpEmitter(remoteAddr, origin)

			udpEmitter.Close()

			err := udpEmitter.Emit(testEvent)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("use of closed network connection"))
		})
	})

	Describe("Emit()", func() {
		var udpEmitter emitter.Emitter

		BeforeEach(func() {
			udpEmitter, _ = emitter.NewUdpEmitter(remoteAddr, origin)
		})

		Context("when the agent is listening", func() {

			var agentListener net.PacketConn

			BeforeEach(func() {
				agentListener, _ = net.ListenPacket("udp", ":12345")
			})

			AfterEach(func() {
				agentListener.Close()
			})

			It("should send the envelope as a []byte", func(done Done) {
				err := udpEmitter.Emit(testEvent)
				Expect(err).ToNot(HaveOccurred())
				buffer := make([]byte, 4096)
				readCount, _, err := agentListener.ReadFrom(buffer)
				Expect(err).ToNot(HaveOccurred())
				var envelope events.Envelope
				err = proto.Unmarshal(buffer[:readCount], &envelope)
				Expect(err).ToNot(HaveOccurred())
				Expect(envelope.GetEventType()).To(Equal(events.Envelope_Heartbeat))
				Expect(envelope.GetHeartbeat()).To(Equal(testEvent))
				Expect(envelope.GetOrigin().GetJobName()).To(Equal(jobName))
				Expect(envelope.GetOrigin().GetJobInstanceId()).To(Equal(jobIndex))
				close(done)
			})
		})

		Context("when the agent is not listening", func() {
			It("should attempt to send the envelope", func() {
				err := udpEmitter.Emit(testEvent)
				Expect(err).ToNot(HaveOccurred())
			})
			Context("then the agent starts Listening", func() {
				It("should eventually send envelopes as a []byte", func(done Done) {
					err := udpEmitter.Emit(testEvent)
					Expect(err).ToNot(HaveOccurred())
					agentListener, err := net.ListenPacket("udp", ":12345")
					Expect(err).ToNot(HaveOccurred())
					err = udpEmitter.Emit(testEvent)
					Expect(err).ToNot(HaveOccurred())
					buffer := make([]byte, 4096)
					readCount, _, err := agentListener.ReadFrom(buffer)
					Expect(err).ToNot(HaveOccurred())
					var envelope events.Envelope
					err = proto.Unmarshal(buffer[:readCount], &envelope)
					Expect(err).ToNot(HaveOccurred())
					Expect(envelope.GetEventType()).To(Equal(events.Envelope_Heartbeat))
					Expect(envelope.GetHeartbeat()).To(Equal(testEvent))
					close(done)
				})
			})
		})
	})

	Describe("NewUdpEmitter()", func() {
		Context("when ResolveUDPAddr fails", func() {
			It("returns an error", func() {
				emitter, err := emitter.NewUdpEmitter("invalid-address:", origin)
				Expect(emitter).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when all is good", func() {
			It("creates an emitter", func() {
				emitter, err := emitter.NewUdpEmitter("localhost:123", origin)
				Expect(emitter).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
