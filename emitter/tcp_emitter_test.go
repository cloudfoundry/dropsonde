package emitter_test

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"net"
)

func createTestListener() (net.Listener, <-chan *events.Envelope) {
	listener, _ := net.Listen("tcp", ":0")
	msgChan := make(chan *events.Envelope)

	go func() {
		defer close(msgChan)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			buf := make([]byte, 4096)
			n, _ := io.ReadFull(conn, buf)

			envelope := &events.Envelope{}
			err = proto.Unmarshal(buf[:n], envelope)
			if err != nil {
				panic(err)
			}

			msgChan <- envelope
		}
	}()

	return listener, msgChan
}

var _ = Describe("TcpEmitter", func() {
	var origin = events.NewOrigin("job-name", 42)

	Describe("NewTcpEmitter()", func() {
		Context("when remoteAddress is parseable", func() {
			It("returns an emitter", func() {
				tcpEmitter, err := emitter.NewTcpEmitter("localhost:123", origin)
				Expect(tcpEmitter).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when remoteAddress is not parseable", func() {
			It("returns an error", func() {
				tcpEmitter, err := emitter.NewTcpEmitter("$#&^bad-address!!!:", origin)
				Expect(tcpEmitter).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Emit()", func() {
		var (
			tcpEmitter emitter.Emitter
			testEvent  = events.NewHeartbeat(123, 0, 0)
		)

		Context("when the agent is listening", func() {
			var (
				testListener net.Listener
				msgChan      <-chan *events.Envelope
			)

			BeforeEach(func() {
				testListener, msgChan = createTestListener()
				tcpEmitter, _ = emitter.NewTcpEmitter(testListener.Addr().String(), origin)
			})

			AfterEach(func() {
				testListener.Close()
			})

			It("sends events to the agent", func(done Done) {
				defer close(done)

				err := tcpEmitter.Emit(testEvent)
				Expect(err).ToNot(HaveOccurred())

				receivedMsg, ok := <-msgChan
				Expect(ok).To(BeTrue())
				Expect(receivedMsg.GetOrigin()).To(Equal(origin))
				Expect(receivedMsg.GetEventType()).To(Equal(events.Envelope_Heartbeat))
				Expect(receivedMsg.GetHeartbeat()).To(Equal(testEvent))
			})
		})

		Context("when the agent is not listening", func() {
			BeforeEach(func() {
				tcpEmitter, _ = emitter.NewTcpEmitter("localhost:123", origin)
			})

			It("returns an error", func() {
				err := tcpEmitter.Emit(testEvent)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
