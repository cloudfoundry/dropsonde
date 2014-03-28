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

	Describe("Emit()", func() {
		var udpEmitter emitter.Emitter
		var testEvent *events.DropsondeStatus

		BeforeEach(func() {
			os.Setenv("BOSH_JOB_NAME", "awesome_job")
			os.Setenv("BOSH_JOB_INSTANCE", "1")
			testEvent = &events.DropsondeStatus{SentCount: proto.Uint64(1), ErrorCount: proto.Uint64(0)}
			udpEmitter, _ = emitter.NewUdpEmitter()
		})

		Context("when the agent is listening", func() {

			var agentListener net.PacketConn

			BeforeEach(func() {
				agentListener, _ = net.ListenPacket("udp", ":42420")
			})

			AfterEach(func() {
				agentListener.Close()
			})

			It("should send the envelope as a []byte", func(done Done) {
				err := udpEmitter.Emit(testEvent)
				Expect(err).To(BeNil())
				buffer := make([]byte, 4096)
				readCount, _, err := agentListener.ReadFrom(buffer)
				Expect(err).To(BeNil())
				var envelope events.Envelope
				err = proto.Unmarshal(buffer[:readCount], &envelope)
				Expect(err).To(BeNil())
				Expect(envelope.GetEventType()).To(Equal(events.Envelope_DropsondeStatus))
				Expect(envelope.GetDropsondeStatus()).To(Equal(testEvent))
				close(done)
			})
		})

		Context("when the agent is not listening", func() {
			It("should attempt to send the envelope", func() {
				err := udpEmitter.Emit(testEvent)
				Expect(err).To(BeNil())
			})
			Context("then the agent starts Listening", func() {
				It("should eventually send envelopes as a []byte", func(done Done) {
					err := udpEmitter.Emit(testEvent)
					Expect(err).To(BeNil())
					agentListener, err := net.ListenPacket("udp", ":42420")
					Expect(err).To(BeNil())
					err = udpEmitter.Emit(testEvent)
					Expect(err).To(BeNil())
					buffer := make([]byte, 4096)
					readCount, _, err := agentListener.ReadFrom(buffer)
					Expect(err).To(BeNil())
					var envelope events.Envelope
					err = proto.Unmarshal(buffer[:readCount], &envelope)
					Expect(err).To(BeNil())
					Expect(envelope.GetEventType()).To(Equal(events.Envelope_DropsondeStatus))
					Expect(envelope.GetDropsondeStatus()).To(Equal(testEvent))
					close(done)
				})
			})
		})
	})

	Describe("NewUdpEmitter()", func() {
		Context("with missing environment variables", func() {
			BeforeEach(func() {
				os.Setenv("BOSH_JOB_NAME", "")
				os.Setenv("BOSH_JOB_INSTANCE", "")
			})

			It("returns an error", func() {
				emitter, err := emitter.NewUdpEmitter()
				Expect(emitter).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("BOSH_JOB_NAME or BOSH_JOB_INSTANCE not set"))
			})
		})

		Context("when ResolveUDPAddr fails", func() {
			var originalDefaultAddress string

			BeforeEach(func() {
				os.Setenv("BOSH_JOB_NAME", "test-job-name")
				os.Setenv("BOSH_JOB_INSTANCE", "0")
				originalDefaultAddress = emitter.DefaultAddress
				emitter.DefaultAddress = "invalid-address:"
			})

			AfterEach(func() {
				emitter.DefaultAddress = originalDefaultAddress
			})

			It("returns an error", func() {
				emitter, err := emitter.NewUdpEmitter()
				Expect(emitter).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})

		Context("when all is good", func() {
			BeforeEach(func() {
				os.Setenv("BOSH_JOB_NAME", "test-job-name")
				os.Setenv("BOSH_JOB_INSTANCE", "0")
			})

			It("creates an emitter", func() {
				emitter, err := emitter.NewUdpEmitter()
				Expect(emitter).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})
	})
})
